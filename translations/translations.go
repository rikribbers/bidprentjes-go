package translations

type Language struct {
	Code string
	Name string
	Flag string
}

var SupportedLanguages = []Language{
	{Code: "en", Name: "English", Flag: "gb"},
	{Code: "nl", Name: "Nederlands", Flag: "nl"},
	{Code: "de", Name: "Deutsch", Flag: "de"},
}

type Translations struct {
	Search               string
	SearchPlaceholder    string
	SearchHelp           string
	Actions              string
	Edit                 string
	Delete               string
	Create               string
	Cancel               string
	Yes                  string
	No                   string
	DeleteConfirm        string
	FirstName            string
	Prefix               string
	LastName             string
	BirthDate            string
	BirthPlace           string
	DeathDate            string
	DeathPlace           string
	Scan                 string
	CreateNew            string
	DeleteError          string
	DeleteSuccess        string
	Upload               string
	SelectCSVFile        string
	CSVFormat            string
	UploadSuccess        string
	UploadError          string
	RecordsImported      string
	CSVFormatDescription string
	Example              string
	CSVDateFormat        string
	CSVScanFormat        string
	CSVHeader            string
	SelectFileError      string
	Uploading            string
	SearchResults        string
	TotalResults         string
	HasScan              string
	NoResults            string
	Page                 string
	Of                   string
	ID                   string
	ExactMatch           string
}

var translations = map[string]Translations{
	"en": {
		Search:               "Search Bidprentjes",
		SearchPlaceholder:    "Search by name, place or year...",
		SearchHelp:           "Search by name, place, or year (e.g., \"Jan Amsterdam 1900\"). Years will match both birth and death years.",
		Actions:              "Actions",
		Edit:                 "Edit",
		Delete:               "Delete",
		Create:               "Create",
		Cancel:               "Cancel",
		Yes:                  "Yes",
		No:                   "No",
		DeleteConfirm:        "Are you sure you want to delete this bidprentje?",
		FirstName:            "First Name",
		Prefix:               "Prefix",
		LastName:             "Last Name",
		BirthDate:            "Birth Date",
		BirthPlace:           "Birth Place",
		DeathDate:            "Death Date",
		DeathPlace:           "Death Place",
		Scan:                 "Scan",
		CreateNew:            "Create New",
		DeleteError:          "Failed to delete bidprentje",
		DeleteSuccess:        "Successfully deleted bidprentje",
		Upload:               "Upload CSV",
		SelectCSVFile:        "Select CSV File",
		CSVFormat:            "CSV should contain: ID, FirstName, Prefix, LastName, BirthDate, BirthPlace, DeathDate, DeathPlace, Scan",
		UploadSuccess:        "Upload successful",
		UploadError:          "Failed to upload file",
		RecordsImported:      "records imported",
		CSVFormatDescription: "The CSV file should have the following columns:",
		Example:              "Example",
		CSVDateFormat:        "Dates should be in YYYY-MM-DD format",
		CSVScanFormat:        "Scan should be either 'true' or 'false'",
		CSVHeader:            "First line should be the header row",
		SelectFileError:      "Please select a file",
		Uploading:            "Uploading...",
		SearchResults:        "Search Results",
		TotalResults:         "Total Results",
		HasScan:              "Has Scan",
		NoResults:            "No results found",
		Page:                 "Page",
		Of:                   "of",
		ID:                   "ID",
		ExactMatch:           "Exact matches only",
	},
	"nl": {
		Search:               "Bidprentjes zoeken",
		SearchPlaceholder:    "Zoek op naam, plaats of jaar...",
		SearchHelp:           "Zoek op naam, plaats of jaar (bijv. \"Jan Amsterdam 1900\"). Jaren worden gezocht in geboorte- en sterfdatum.",
		Actions:              "Acties",
		Edit:                 "Bewerken",
		Delete:               "Verwijderen",
		Create:               "Aanmaken",
		Cancel:               "Annuleren",
		Yes:                  "Ja",
		No:                   "Nee",
		DeleteConfirm:        "Weet u zeker dat u dit bidprentje wilt verwijderen?",
		FirstName:            "Voornaam",
		Prefix:               "Tussenvoegsel",
		LastName:             "Achternaam",
		BirthDate:            "Geboortedatum",
		BirthPlace:           "Geboorteplaats",
		DeathDate:            "Overlijdensdatum",
		DeathPlace:           "Overlijdensplaats",
		Scan:                 "Scan",
		CreateNew:            "Nieuw",
		DeleteError:          "Fout bij verwijderen bidprentje",
		DeleteSuccess:        "Bidprentje succesvol verwijderd",
		Upload:               "CSV Uploaden",
		SelectCSVFile:        "Selecteer CSV Bestand",
		CSVFormat:            "CSV moet bevatten: ID, Voornaam, Tussenvoegsel, Achternaam, Geboortedatum, Geboorteplaats, Overlijdensdatum, Overlijdensplaats, Scan",
		UploadSuccess:        "Upload succesvol",
		UploadError:          "Fout bij uploaden bestand",
		RecordsImported:      "records geïmporteerd",
		CSVFormatDescription: "Het CSV-bestand moet de volgende kolommen bevatten:",
		Example:              "Voorbeeld",
		CSVDateFormat:        "Datums moeten in JJJJ-MM-DD formaat zijn",
		CSVScanFormat:        "Scan moet 'true' of 'false' zijn",
		CSVHeader:            "Eerste regel moet de kolomnamen bevatten",
		SelectFileError:      "Selecteer een bestand",
		Uploading:            "Uploaden...",
		SearchResults:        "Zoekresultaten",
		TotalResults:         "Totaal aantal resultaten",
		HasScan:              "Scan beschikbaar",
		NoResults:            "Geen resultaten gevonden",
		Page:                 "Pagina",
		Of:                   "van",
		ID:                   "ID",
		ExactMatch:           "Alleen exacte overeenkomsten",
	},
	"de": {
		Search:               "Bidprentjes suchen",
		SearchPlaceholder:    "Nach Name, Ort oder Jahr suchen...",
		SearchHelp:           "Suche nach Name, Ort oder Jahr (z.B. \"Jan Amsterdam 1900\"). Jahre werden in Geburts- und Sterbedatum gesucht.",
		Actions:              "Aktionen",
		Edit:                 "Bearbeiten",
		Delete:               "Löschen",
		Create:               "Erstellen",
		Cancel:               "Abbrechen",
		Yes:                  "Ja",
		No:                   "Nein",
		DeleteConfirm:        "Möchten Sie dieses Bidprentje wirklich löschen?",
		FirstName:            "Vorname",
		Prefix:               "Präfix",
		LastName:             "Nachname",
		BirthDate:            "Geburtsdatum",
		BirthPlace:           "Geburtsort",
		DeathDate:            "Sterbedatum",
		DeathPlace:           "Sterbeort",
		Scan:                 "Scan",
		CreateNew:            "Neu",
		DeleteError:          "Fehler beim Löschen des Bidprentje",
		DeleteSuccess:        "Bidprentje erfolgreich gelöscht",
		Upload:               "CSV hochladen",
		SelectCSVFile:        "CSV-Datei auswählen",
		CSVFormat:            "CSV muss enthalten: ID, Vorname, Präfix, Nachname, Geburtsdatum, Geburtsort, Sterbedatum, Sterbeort, Scan",
		UploadSuccess:        "Upload erfolgreich",
		UploadError:          "Fehler beim Hochladen der Datei",
		RecordsImported:      "Datensätze importiert",
		CSVFormatDescription: "Die CSV-Datei muss folgende Spalten enthalten:",
		Example:              "Beispiel",
		CSVDateFormat:        "Daten müssen im JJJJ-MM-TT Format sein",
		CSVScanFormat:        "Scan muss 'true' oder 'false' sein",
		CSVHeader:            "Erste Zeile muss die Spaltenüberschriften enthalten",
		SelectFileError:      "Bitte wählen Sie eine Datei aus",
		Uploading:            "Wird hochgeladen...",
		SearchResults:        "Suchergebnisse",
		TotalResults:         "Gesamtergebnisse",
		HasScan:              "Scan verfügbar",
		NoResults:            "Keine Ergebnisse gefunden",
		Page:                 "Seite",
		Of:                   "von",
		ID:                   "ID",
		ExactMatch:           "Nur exakte Übereinstimmungen",
	},
}

func GetTranslation(lang string) Translations {
	if t, ok := translations[lang]; ok {
		return t
	}
	return translations["nl"] // fallback to Dutch instead of English
}
