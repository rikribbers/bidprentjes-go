import csv
import random
from datetime import datetime, timedelta

# Sample data
voornamen = [
    # Traditional Dutch names
    "Jan", "Piet", "Klaas", "Marie", "Anna", "Willem", "Hendrik", "Elisabeth", "Johannes", "Cornelia",
    "Gerrit", "Johanna", "Petrus", "Maria", "Cornelis", "Adrianus", "Wilhelmus", "Antonia", "Theodorus", "Helena",
    "Jacobus", "Catharina", "Franciscus", "Margaretha", "Antonius", "Christina", "Bernardus", "Geertruida", "Martinus", "Hendrika",
    "Albertus", "Alida", "Dirk", "Emma", "Frederik", "Grietje", "Hermanus", "Ida", "Joseph", "Klasina",
    "Lambertus", "Martha", "Nicolaas", "Neeltje", "Otto", "Pieternella", "Quirinus", "Rosa", "Simon", "Sophia",
    "Thomas", "Trijntje", "Ubbo", "Ursula", "Vincent", "Willemina", "Xavier", "Yda", "Zacharias", "Zwaantje",
    # Modern Dutch names
    "Sem", "Lucas", "Levi", "Finn", "Noah", "Daan", "Milan", "Liam", "James", "Luuk",
    "Emma", "Julia", "Mila", "Sophie", "Tess", "Sara", "Nova", "Nora", "Liv", "Zoë",
    "Bram", "Lars", "Jesse", "Benjamin", "Thomas", "Sam", "Thijs", "Adam", "Max", "Julian",
    "Sophie", "Eva", "Maud", "Luna", "Lotte", "Nina", "Milou", "Evi", "Saar", "Roos",
    # Additional traditional names
    "Adriaan", "Bart", "Christiaan", "David", "Eduard", "Frans", "Gerard", "Herman", "Izaak", "Jacob",
    "Karel", "Leo", "Maarten", "Nico", "Oscar", "Paul", "Quintijn", "Rudolf", "Stefan", "Theo",
    # Additional modern names
    "Aiden", "Boaz", "Cas", "Dex", "Elias", "Floris", "Gijs", "Hugo", "Ian", "Jayden",
    "Kai", "Luca", "Mason", "Noud", "Owen", "Pim", "Quinn", "Ruben", "Stijn", "Thijmen",
    # Female names
    "Amber", "Bente", "Charlotte", "Daphne", "Eline", "Femke", "Guusje", "Hanna", "Iris", "Jasmijn",
    "Kim", "Laura", "Marit", "Nienke", "Olivia", "Puck", "Quinty", "Robin", "Sarah", "Tessa"
]

tussenvoegsels = ["van", "de", "van der", "van den", "", "ter", "van de", "", "van", "de", 
                  "den", "der", "'t", "ten", "te", "op de", "bij de", "aan de", "in de", "onder de"]

achternamen = [
    "Berg", "Vries", "Bakker", "Janssen", "Visser", "Molen", "Bosch", "Groot", "Klein", "Smit",
    "Hendriks", "Peters", "Dekker", "Brouwer", "Mulder", "Meyer", "Dijkstra", "Post", "Hoekstra", "Kok",
    "Jansen", "de Jong", "van Dijk", "Bakker", "de Vries", "van den Berg", "van der Meer", "de Boer", "Prins", "Mulder",
    "Veenstra", "Kramer", "van Leeuwen", "Scholten", "van Wijk", "Postma", "Martens", "Vos", "de Graaf", "Mol"
]

plaatsen = [
    # Major cities
    "Amsterdam", "Rotterdam", "Den Haag", "Utrecht", "Eindhoven", "Groningen", "Tilburg", "Almere", "Breda", "Nijmegen",
    "Enschede", "Haarlem", "Arnhem", "Zaanstad", "Amersfoort", "Apeldoorn", "Hoofddorp", "Maastricht", "Leiden", "Dordrecht",
    # Medium cities
    "Zoetermeer", "Zwolle", "Emmen", "Delft", "Venlo", "Deventer", "Alkmaar", "Helmond", "Hilversum", "Ede",
    "Roosendaal", "Oosterhout", "Hengelo", "Purmerend", "Schiedam", "Spijkenisse", "Leeuwarden", "Gouda", "Bergen op Zoom", "Alphen aan den Rijn",
    # Smaller cities and towns
    "Assen", "Veenendaal", "Zeist", "Hoorn", "Middelburg", "Kampen", "Weert", "Woerden", "Boxtel", "Bussum",
    "Huizen", "Winterswijk", "Sneek", "Waalwijk", "Tiel", "Vlaardingen", "Meppel", "Oldenzaal", "Sittard", "Roermond",
    # Historical towns
    "Naarden", "Muiden", "Enkhuizen", "Edam", "Volendam", "Zutphen", "Doesburg", "Vianen", "Wijk bij Duurstede", "Coevorden",
    "Franeker", "Workum", "Sloten", "Stavoren", "IJsselstein", "Montfoort", "Nieuwpoort", "Schoonhoven", "Brielle", "Geervliet",
    # Northern cities
    "Delfzijl", "Winschoten", "Veendam", "Stadskanaal", "Hoogezand", "Harlingen", "Bolsward", "Dokkum", "Drachten", "Heerenveen",
    # Southern cities
    "Terneuzen", "Goes", "Vlissingen", "Hulst", "Oostburg", "Zierikzee", "Tholen", "Bergen op Zoom", "Roosendaal", "Waalwijk",
    # Eastern cities
    "Almelo", "Hardenberg", "Rijssen", "Denekamp", "Ootmarsum", "Oldenzaal", "Borne", "Goor", "Lochem", "Borculo",
    # Western cities
    "Katwijk", "Noordwijk", "Lisse", "Hillegom", "Aalsmeer", "Uithoorn", "Mijdrecht", "Bodegraven", "Waddinxveen", "Ridderkerk"
]

def generate_birth_date():
    start_date = datetime(1900, 1, 1)
    end_date = datetime(1950, 12, 31)
    days_between = (end_date - start_date).days
    random_days = random.randint(0, days_between)
    return start_date + timedelta(days=random_days)

def generate_death_date(birth_date):
    min_age = 60
    max_age = 100
    age = random.randint(min_age, max_age)
    return birth_date + timedelta(days=age*365)

# Create CSV file
with open('test_data.csv', 'w', newline='') as f:
    writer = csv.writer(f)
    
    # Write header
    writer.writerow(['id', 'voornaam', 'tussenvoegsel', 'achternaam', 'geboortedatum', 
                    'geboorteplaats', 'overlijdensdatum', 'overlijdensplaats', 'scan'])
    
    # Generate 10000 entries
    for i in range(1, 10001):  # Start from 1 to 10000
        birth_date = generate_birth_date()
        death_date = generate_death_date(birth_date)
        
        row = [
            str(i),  # Use sequential number as ID
            random.choice(voornamen),
            random.choice(tussenvoegsels),
            random.choice(achternamen),
            birth_date.strftime('%Y-%m-%d'),
            random.choice(plaatsen),
            death_date.strftime('%Y-%m-%d'),
            random.choice(plaatsen),
            str(random.choice([True, False])).lower()
        ]
        
        writer.writerow(row)

print("Generated test_data.csv with 10,000 entries") 