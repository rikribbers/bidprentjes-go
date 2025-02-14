I want to build an API that can handle the managing of a simple model of a "Bidprentje" object in the go programming language.

The Bidprentje object has the following properties:
- id: string
- voornaam: string
- tussenvoegsel: string
- achternaam: string
- geboortedatum: date
- geboorteplaats: string
- overlijdensdatum: date
- overlijdensplaats: string
- scan: boolean
- created_at: date
- updated_at: date

The API Should handle the following operations:
- Create a new Bidprentje
- Get a Bidprentje by ID
- Update an existing Bidprentje
- Delete a Bidprentje by ID
- List all Bidprentjes
- Search Bidprentjes

"Search Bidprentjes" should be able to do a fuzzy search on all fields. 


# API Examples

## Create

````bash
curl -X POST \
http://localhost:8080/bidprentjes \
-H 'Content-Type: application/json' \
-d '{
"voornaam": "Jan",
"tussenvoegsel": "van",
"achternaam": "Berg",
"geboortedatum": "1900-01-01T00:00:00Z",
"geboorteplaats": "Amsterdam",
"overlijdensdatum": "1980-12-31T00:00:00Z",
"overlijdensplaats": "Rotterdam",
"scan": true
}'
````

## Get by ID

````bash
curl -X GET http://localhost:8080/bidprentjes/550e8400-e29b-41d4-a716-446655440000
````

## Update

````bash
curl -X PUT \
http://localhost:8080/bidprentjes/550e8400-e29b-41d4-a716-446655440000 \
-H 'Content-Type: application/json' \
-d '{
"voornaam": "Jan",
"tussenvoegsel": "van",
"achternaam": "Berg",
"geboortedatum": "1900-01-01T00:00:00Z",
"geboorteplaats": "Utrecht",
"overlijdensdatum": "1980-12-31T00:00:00Z",
"overlijdensplaats": "Rotterdam",
"scan": true
}'

## Delete

````bash
curl -X DELETE http://localhost:8080/bidprentjes/550e8400-e29b-41d4-a716-446655440000
````

## List all

````bash
curl -X GET http://localhost:8080/bidprentjes
````

## Search

````bash
curl -X POST \
http://localhost:8080/bidprentjes/search \
-H 'Content-Type: application/json' \
-d '{"query": "amsterdam"}'
````

# Generate CSV

````bash
python3 scripts/generate_csv.py
````