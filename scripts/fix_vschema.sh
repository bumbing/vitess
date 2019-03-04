#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# If for whatever reason ads-latest has bad vschemas,
# this script attempts to forcefully clean it up by
# adding any missing sequence tables and updating the
# vschemas in latest. The sequence tables need to be
# created in prod patio, though, or else they'll disappear
# from ads-latest at night during the dump from prod.

PATIO_ARGS=""
PATIOGENERAL_ARGS=""
UPDATE_GENERAL=true

VTENV="$1"
if [[ "$VTENV" == "test" || "$VTENV" == "latest" ]]; then
  PATIO_ARGS=(-create-sequences -include-cols -cols-authoritative
              -create-primary-vindexes -create-secondary-vindexes
              -default-scatter-cache-capacity 100000)
  PATIOGENERAL_ARGS=(-include-cols -cols-authoritative)
elif [[ "$VTENV" == "shadow" ]]; then
  PATIO_ARGS=(-include-cols -cols-authoritative
              -create-primary-vindexes -create-secondary-vindexes
              -default-scatter-cache-capacity 100000
              -table-scatter-cache-capacity "campaigns:200000,product_groups:1000000"
             )
  PATIOGENERAL_ARGS=(-include-cols -cols-authoritative)
  UPDATE_GENERAL=false
elif [[ "$VTENV" == "prod" ]]; then
  # TODO(dweitzman): Start rolling out sequences by adding
  # "-seq-table-whitelist=accepted_tos" to the patio arguments.
  PATIO_ARGS=(-include-cols)
  PATIOGENERAL_ARGS=(-include-cols)
else
  echo "Unsupported env name: $1"
  exit 1
fi

echo Operating on environment "$VTENV". For ads-latest, run "$0" latest

PVCTL_CMD="/vt/scripts/pvtctl.sh"
PINSCHEMA_CMD="/vt/bin/pinschema"
if [[ "$OSTYPE" == "darwin"* ]]; then
  if command -v pinschema && command -v pvtctl.sh; then
    # This person probably installed https://phabricator.pinadmin.com/diffusion/BREW/
    PVCTL_CMD="pvtctl.sh"
    PINSCHEMA_CMD="pinschema"
  else
    # This person probably did not install vitess from homebrew. Make a best effort attempt
    # to compile on demand
    echo "In darwin/laptop mode without vitess binaries in \$PATH

    Consider installing them with:
      \$ brew tap pinterest/brewpub ssh://git@phabricator.pinadmin.com/diffusion/BREW/brewpub.git
      \$ brew tap-pin pinterest/brewpub
      \$ brew install vitess

     For now we're trying to compile on-demand. You'll need:
      go >=1.11
      ~/code/vitess/go.mod created by vt_fix_gomod.sh"
    PVCTL_CMD="./scripts/pvtctl.sh"
    PINSCHEMA_CMD="go run ./go/cmd/pinschema"
  fi
fi

echo Validating consistent shard schemas...
$PVCTL_CMD "$VTENV" ValidateSchemaKeyspace patio

echo Finding tablets to pull schemas from...
PATIO_MASTER=$($PVCTL_CMD "$VTENV" ListAllTablets | grep " patio 0 " | grep " master " | cut -d' ' -f 1)
if $UPDATE_GENERAL; then
  PATIOGENERAL_MASTER=$($PVCTL_CMD "$VTENV" ListAllTablets | grep " patiogeneral 0 " | grep " master " | cut -d' ' -f 1)
fi

echo Saving current patio schema...
PATIO_SCHEMA_FILE=$(mktemp -t patio-schema.sql.XXXX)
PATIO_SCHEMA_CONTENT=$($PVCTL_CMD "$VTENV" GetSchema "$PATIO_MASTER" | jq -r '.table_definitions[].schema + ";"')
echo "$PATIO_SCHEMA_CONTENT" > "$PATIO_SCHEMA_FILE"

if [[ "${ADD_SEQS:-false}" == "true" ]]; then
    echo Making sure sequence tables exist in patiogeneral...
    CREATE_SQL=$($PINSCHEMA_CMD create-seq "$PATIO_SCHEMA_FILE")
    $PVCTL_CMD "$VTENV" ApplySchema -sql="$CREATE_SQL" patiogeneral
fi

if $UPDATE_GENERAL; then
  echo Saving current patiogeneral schema...
  PATIOGENERAL_SCHEMA_FILE=$(mktemp -t patiogeneral-schema.sql.XXXX)
  PATIOGENERAL_SCHEMA_CONTENT=$($PVCTL_CMD "$VTENV" GetSchema "$PATIOGENERAL_MASTER" | jq -r '.table_definitions[].schema + ";"')
  echo "$PATIOGENERAL_SCHEMA_CONTENT" > "$PATIOGENERAL_SCHEMA_FILE"

  echo Diffing patiogeneral vschema...
  PATIOGENERAL_VSCHEMA=$($PINSCHEMA_CMD create-vschema "${PATIOGENERAL_ARGS[@]}" "$PATIOGENERAL_SCHEMA_FILE")
  PATIOGENERAL_VSCHEMA_OLD=$($PVCTL_CMD "$VTENV" GetVSchema patiogeneral)

  DIFF=$(diff --strip-trailing-cr -u <(echo "$PATIOGENERAL_VSCHEMA_OLD") <(echo "$PATIOGENERAL_VSCHEMA") || test "$?" -eq 1)
  if [ "$DIFF" ]; then
    echo "$DIFF"

    while read -p "Does this change to patiogeneral vschema look right (y/n)? " choice
    do
      case "$choice" in
        y|Y|n|N ) break;;
        * ) echo "Type 'y' or 'n'"; continue ;;
      esac
    done

    case "$choice" in
      y|Y ) $PVCTL_CMD "$VTENV" ApplyVSchema -vschema="$PATIOGENERAL_VSCHEMA" patiogeneral;;
      * ) echo "Cancelled"; exit 1;;
    esac
  else
    echo "No change to patiogeneral vschema"
  fi
fi

echo Diffing patio vschema...
PATIO_VSCHEMA=$($PINSCHEMA_CMD create-vschema "${PATIO_ARGS[@]}" "$PATIO_SCHEMA_FILE")
PATIO_VSCHEMA_OLD=$($PVCTL_CMD "$VTENV" GetVSchema patio)
DIFF=$(diff --strip-trailing-cr -u <(echo "$PATIO_VSCHEMA_OLD") <(echo "$PATIO_VSCHEMA") || test "$?" -eq 1)
if [ "$DIFF" ]; then
  echo "$DIFF"

  while read -p "Does this change to patio vschema look right (y/n)? " choice
  do
    case "$choice" in
      y|Y|n|N ) break;;
      * ) echo "Type 'y' or 'n'"; continue ;;
    esac
  done

  case "$choice" in
    y|Y ) $PVCTL_CMD "$VTENV" ApplyVSchema -vschema="$PATIO_VSCHEMA" patio;;
    * ) echo "Cancelled"; exit 1;;
  esac
else
  echo "No change to patio vschema"
fi

echo Cleaning up temp files...
rm -f "$PATIO_SCHEMA_FILE"
if $UPDATE_GENERAL; then
  rm -f "$PATIOGENERAL_SCHEMA_FILE"
fi
