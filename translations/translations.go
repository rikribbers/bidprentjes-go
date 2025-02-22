package translations

type Language struct {
	Code string
	Name string
	Flag string
}

var SupportedLanguages = []Language{
	{Code: "en", Name: "English", Flag: "gb"},
	{Code: "nl", Name: "Nederlands", Flag: "nl"},
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
	},
	"nl": {
		Search:               "Zoek Bidprentjes",
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
		CreateNew:            "Nieuw Aanmaken",
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
	},
}

func GetTranslation(lang string) Translations {
	if t, ok := translations[lang]; ok {
		return t
	}
	return translations["en"] // fallback to English
}
