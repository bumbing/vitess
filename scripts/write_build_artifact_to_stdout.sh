#!/bin/bash
cd /vt/ && (tar cf - ./ | gzip -f -)
