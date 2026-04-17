#!/usr/bin/env bash

# Configuration
BUCKET_NAME="cdn.rikribbers.nl"
OUTPUT="output.csv"

# Write CSV header
echo "identifier,guid,filename" > "$OUTPUT"

# List all objects in the bucket
gsutil ls "gs://${BUCKET_NAME}" | while read -r file; do
    # Skip folder-like entries (ending with /)
    if [[ "$file" == */ ]]; then
        continue
    fi

    filename=$(basename "$file")
    identifier=""

    ###########################################################
    # 1. Special case: filename starts with "ID <number>"
    #    Examples:
    #      "ID 264 Abbink..."
    #      "id   77A something..."
    #    We allow extra text after the number.
    ###########################################################
    if [[ "$filename" =~ ^[Ii][Dd][[:space:]]+([0-9]+)[A-Ba-b]? ]]; then
        identifier="${BASH_REMATCH[1]}"
    else
        ###########################################################
        # 2. General case: standalone ID<number>[A|B] anywhere
        #    Examples:
        #      "scan ID123A final.jpg"
        #      "foo-ID7b-bar.png"
        #      "ID0045_image.png"
        ###########################################################
        raw_identifier=$(echo "$filename" | grep -ioE '\bID[0-9]+[A-B]?\b')

        if [[ -n "$raw_identifier" ]]; then
            # Strip "ID" prefix (case-insensitive)
            id_no_prefix=$(echo "$raw_identifier" | sed 's/^[Ii][Dd]//')

            # Strip trailing A/B variant (case-insensitive)
            identifier=$(echo "$id_no_prefix" | sed 's/[A-Ba-b]$//')
        fi
    fi

    # Generate a GUID using kernel source
    guid=$(cat /proc/sys/kernel/random/uuid)

    # Append to CSV
    echo "${identifier},${guid},${filename}" >> "$OUTPUT"
done

echo "CSV generated: $OUTPUT"