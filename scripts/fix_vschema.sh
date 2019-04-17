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
SKIP_VALIDATE="${SKIP_VALIDATE:-false}"
LOOKUP_VINDEX_WHITELIST='-lookup-vindex-whitelist \"accepted_tos_id_idx,ad_group_spec_id_idx,ad_group_id_idx,advertiser_id_idx,app_event_tracking_config_id_idx,bill_detail_id_idx,billing_action_id_idx,billing_contact_id_idx,billing_profile_id_idx,bill_id_idx,bulk_v2_job_id_idx,business_profile_id_idx,campaign_spec_id_idx,campaign_id_idx,carousel_slot_promotion_id_idx,conversion_tag_id_idx,goal_id_idx,notification_id_idx,order_line_spec_id_idx,order_line_id_idx,pin_promotion_spec_id_idx,pin_promotion_id_idx,pinner_list_spec_id_idx,pinner_list_id_idx,product_group_spec_id_idx,product_group_id_idx,promoted_catalog_product_group_id_idx,promoted_catalog_product_groups_history_id_idx,rule_subscription_id_idx,targeting_attribute_history_id_idx,targeting_attribute_id_idx,targeting_spec_id_idx,user_preference_id_idx\" '

VSCHEMA_ROLLBACK=""
if [[ $# -gt 1 ]]; then
  VSCHEMA_ROLLBACK="$2"
fi

VTENV="$1"
if [[ "$VTENV" == "test" || "$VTENV" == "latest" ]]; then
  PATIO_ARGS=(-create-sequences -include-cols -cols-authoritative
              -create-primary-vindexes -create-secondary-vindexes
              -default-scatter-cache-capacity 100000
              -validate-keyspace patio -validate-shards 2
              "$LOOKUP_VINDEX_WHITELIST"
              )
  PATIOGENERAL_ARGS=(-include-cols -cols-authoritative)

  # TODO(dweitzman): Remove this after turning down the old shard that still has
  # autoincrement.
  SKIP_VALIDATE="true"
elif [[ "$VTENV" == "shadow" ]]; then
  PATIO_ARGS=(-include-cols -cols-authoritative
              -create-primary-vindexes -create-secondary-vindexes
              -default-scatter-cache-capacity 100000
              -table-scatter-cache-capacity "campaigns:200000,product_groups:1000000"
              -validate-keyspace patio -validate-shards 2
              "$LOOKUP_VINDEX_WHITELIST"
             )
  PATIOGENERAL_ARGS=(-include-cols -cols-authoritative)
  UPDATE_GENERAL=false
elif [[ "$VTENV" == "prod" ]]; then
  PATIO_ARGS=(
    -include-cols -cols-authoritative -create-sequences
    -create-primary-vindexes -create-secondary-vindexes
    -default-scatter-cache-capacity 100000
    -table-scatter-cache-capacity "campaigns:200000,product_groups:1000000"
    -validate-keyspace patio -validate-shards 2
    "$LOOKUP_VINDEX_WHITELIST"
  )
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
      ~/code/vitess/go.mod created by ./scripts/fix_gomod.sh"
    PVCTL_CMD="./scripts/pvtctl.sh"
    PINSCHEMA_CMD="go run ./go/cmd/pinschema"
  fi
fi

if [[ "$VSCHEMA_ROLLBACK" == "--rollback" ]]; then
  if [[ $# -lt 3 ]]; then
    echo "rollback must provide golden md5 as: $0 $1 $2 {md5}"
    exit 1
  else
    GOLDEN_MD5="$3"
    VSCHEMA_GOLDEN_FILE="/vt/vtdataroot/vschema_${GOLDEN_MD5}.backup"
    if ! $PVCTL_CMD "$VTENV" ApplyVSchema -vschema_file="$VSCHEMA_GOLDEN_FILE" patio; then
      echo "rollback failed, please double check the given golden md5: ${GOLDEN_MD5}"
      exit 1
    else
      echo "rollback succeeded with the given golden md5: ${GOLDEN_MD5}"
    fi
  fi
  exit 0
fi

if [[ "$SKIP_VALIDATE" != "true" ]]; then
  echo Validating consistent shard schemas...
  $PVCTL_CMD "$VTENV" ValidateSchemaKeyspace patio
fi

echo Finding tablets to pull schemas from...
PATIO_MASTER=$($PVCTL_CMD "$VTENV" ListAllTablets | grep " patio -80 " | grep " master " | cut -d' ' -f 1)
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
    echo "$DIFF" | less

    while read -rp "Does this change to patiogeneral vschema look right (y/n)? " choice
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

echo Validating patio vschema ...
PATIO_VSCHEMA=$($PINSCHEMA_CMD create-vschema "${PATIO_ARGS[@]}" "$PATIO_SCHEMA_FILE")
PATIO_VSCHEMA_FILE=$(mktemp -t patio-vschema.json.XXXX)
echo "$PATIO_VSCHEMA" > "$PATIO_VSCHEMA_FILE"

if VALIDATION=$($PINSCHEMA_CMD validate-vschema "${PATIO_ARGS[@]}" -validate-vschema-file "$PATIO_VSCHEMA_FILE" "$PATIO_SCHEMA_FILE" 2>&1); then
  PATIO_VSCHEMA_OLD=$($PVCTL_CMD "$VTENV" GetVSchema patio)
  DIFF=$(diff --strip-trailing-cr -u <(echo "$PATIO_VSCHEMA_OLD") <(echo "$PATIO_VSCHEMA") || test "$?" -eq 1)
  if [ "$DIFF" ]; then
    # backup the vschema golden file to vtctld:/mnt/vtdataroot/
    if [[ -z "$(command -v md5sum)" ]]; then
      GOLDEN_MD5=$(echo -n "${PATIO_VSCHEMA_OLD}" | md5)
    else
      GOLDEN_MD5=$(echo -n "${PATIO_VSCHEMA_OLD}" | md5sum)
      GOLDEN_MD5="${GOLDEN_MD5%  -}"
    fi
    VTCTLD_HOST=$(fh -h "vtctld-${VTENV}" | head -n 1)
    VSCHEMA_GOLDEN_FILE="/mnt/vtdataroot/vschema_${GOLDEN_MD5}.backup"
    echo "${PATIO_VSCHEMA_OLD}" | ssh -T "$VTCTLD_HOST" "sudo tee ${VSCHEMA_GOLDEN_FILE} >/dev/null"

    echo "$DIFF" | less

    while read -rp "Does this change to patio vschema look right (y/n)? " choice
    do
      case "$choice" in
        y|Y|n|N ) break;;
        * ) echo "Type 'y' or 'n'"; continue ;;
      esac
    done

    case "$choice" in
      y|Y )
        $PVCTL_CMD "$VTENV" ApplyVSchema -vschema="$PATIO_VSCHEMA" patio
        echo "VSchema applied, if rollback needed, please use: $0 $1 --rollback ${GOLDEN_MD5}"
      ;;
      * ) echo "Cancelled"; exit 1;;
    esac
  else
    echo "No change to patio vschema"
  fi
else
  echo "Validate patio vschema failed, because of: ${VALIDATION}"
fi

echo Cleaning up temp files...
rm -f "$PATIO_SCHEMA_FILE"
rm -f "$PATIO_VSCHEMA_FILE"
if $UPDATE_GENERAL; then
  rm -f "$PATIOGENERAL_SCHEMA_FILE"
fi
