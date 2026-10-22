#!/bin/sh
# Creates every topic in libs/events/topics.txt plus its .retry and .dlq companions.
# Idempotent (--if-not-exists), so it is safe on every `make up`.
set -eu

BIN=/opt/kafka/bin/kafka-topics.sh
count=0
while IFS= read -r topic || [ -n "$topic" ]; do
  [ -z "$topic" ] && continue
  for t in "$topic" "$topic.retry" "$topic.dlq"; do
    # DLQs keep messages for 14 days for manual inspection and replay.
    retention=604800000
    case "$t" in *.dlq) retention=1209600000 ;; esac
    "$BIN" --bootstrap-server "$BOOTSTRAP" --create --if-not-exists \
      --topic "$t" --partitions "$PARTITIONS" --replication-factor 1 \
      --config retention.ms="$retention" >/dev/null
    count=$((count + 1))
  done
done < /scripts/topics.txt
echo "kafka-init: ensured $count topics"
"$BIN" --bootstrap-server "$BOOTSTRAP" --list
