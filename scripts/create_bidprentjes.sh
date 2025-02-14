#!/bin/bash

# Array of sample data
voornamen=("Jan" "Piet" "Klaas" "Marie" "Anna" "Willem" "Hendrik" "Elisabeth" "Johannes" "Cornelia")
tussenvoegsels=("van" "de" "van der" "van den" "" "ter" "van de" "" "van" "de")
achternamen=("Berg" "Vries" "Bakker" "Janssen" "Visser" "Molen" "Bosch" "Groot" "Klein" "Smit")
plaatsen=("Amsterdam" "Rotterdam" "Utrecht" "Den Haag" "Eindhoven" "Groningen" "Tilburg" "Almere" "Breda" "Nijmegen")

# Function to generate a random date between 1900 and 1950
generate_birth_date() {
    year=$(( ( RANDOM % 50 ) + 1900 ))
    month=$(( ( RANDOM % 12 ) + 1 ))
    day=$(( ( RANDOM % 28 ) + 1 ))
    printf "%04d-%02d-%02dT00:00:00Z" $year $month $day
}

# Function to generate a death date based on birth date
generate_death_date() {
    birth_year=$1
    death_year=$(( birth_year + ( RANDOM % 90 ) + 60 ))  # Person lives 60-150 years
    month=$(( ( RANDOM % 12 ) + 1 ))
    day=$(( ( RANDOM % 28 ) + 1 ))
    printf "%04d-%02d-%02dT00:00:00Z" $death_year $month $day
}

# Create 100 Bidprentjes
for i in {1..100}; do
    birth_date=$(generate_birth_date)
    birth_year=${birth_date:0:4}
    death_date=$(generate_death_date $birth_year)
    
    # Get random indices for the arrays
    voornaam_idx=$(( RANDOM % ${#voornamen[@]} ))
    tussen_idx=$(( RANDOM % ${#tussenvoegsels[@]} ))
    achternaam_idx=$(( RANDOM % ${#achternamen[@]} ))
    geboorte_idx=$(( RANDOM % ${#plaatsen[@]} ))
    overlijden_idx=$(( RANDOM % ${#plaatsen[@]} ))
    
    # Create JSON payload
    json_data=$(cat <<EOF
{
    "voornaam": "${voornamen[$voornaam_idx]}",
    "tussenvoegsel": "${tussenvoegsels[$tussen_idx]}",
    "achternaam": "${achternamen[$achternaam_idx]}",
    "geboortedatum": "$birth_date",
    "geboorteplaats": "${plaatsen[$geboorte_idx]}",
    "overlijdensdatum": "$death_date",
    "overlijdensplaats": "${plaatsen[$overlijden_idx]}",
    "scan": $([ $(( RANDOM % 2 )) -eq 0 ] && echo "false" || echo "true")
}
EOF
    )
    
    # Send POST request
    echo "Creating Bidprentje $i of 100: ${voornamen[$voornaam_idx]} ${tussenvoegsels[$tussen_idx]} ${achternamen[$achternaam_idx]}"
    curl -s -X POST \
         -H "Content-Type: application/json" \
         -d "$json_data" \
         http://localhost:8080/bidprentjes > /dev/null
    
    echo "Done"
done

# List total count
echo -e "\nListing total count of Bidprentjes:"
curl -s -X GET http://localhost:8080/bidprentjes | jq length 