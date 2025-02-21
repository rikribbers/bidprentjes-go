I want to build an application that can handle the managing of a simple model of a "Bidprentje" object in the go programming language.

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

The data should be stored in an fast in memory data store. with the option to store the data
in a file.
The data should be indexed for fast fuzzy search, and all string fields and the date fields are indexed.
The id, created_at and updated_at fields should not be indexed.

For displaying the model I want to create three different web pages:

* /web for editing the model with standard CRUD operations. make it paginated
* /search for searching the model, the result should be displayed paginated with 25 results per page
* /upload for uploading a csv file with bidprentjes data and importing it into the in-memory store; Make sure that the id field is used as the unique identifier for each record.

Each web interface should be behind a different path on the url.

Add a script for generating example data for the module. I should generate 10000 datapoints.
Add a Makefile for building and running the application.
Add a Dockerfile for building the application.