import random
from datetime import datetime, timedelta

# Sample data
voornamen = ["Jan", "Piet", "Klaas", "Marie", "Anna", "Willem", "Hendrik", "Elisabeth", "Johannes", "Cornelia",
             "Gerrit", "Johanna", "Petrus", "Maria", "Cornelis", "Adrianus", "Wilhelmus", "Antonia", "Theodorus", "Helena"]
tussenvoegsels = ["van", "de", "van der", "van den", "", "ter", "van de", "", "van", "de"]
achternamen = ["Berg", "Vries", "Bakker", "Janssen", "Visser", "Molen", "Bosch", "Groot", "Klein", "Smit",
               "Hendriks", "Peters", "Dekker", "Brouwer", "Mulder", "Meyer", "Dijkstra", "Post", "Hoekstra", "Kok"]
plaatsen = ["Amsterdam", "Rotterdam", "Utrecht", "Den Haag", "Eindhoven", "Groningen", "Tilburg", "Almere", "Breda", "Nijmegen",
            "Leiden", "Delft", "Haarlem", "Arnhem", "Maastricht", "Zwolle", "Dordrecht", "Enschede", "Amersfoort", "Zaandam"]

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
with open('sample_bidprentjes.csv', 'w') as f:
    # Write header
    f.write('voornaam,tussenvoegsel,achternaam,geboortedatum,geboorteplaats,overlijdensdatum,overlijdensplaats,scan\n')
    
    # Generate 1000 entries
    for _ in range(1000):
        birth_date = generate_birth_date()
        death_date = generate_death_date(birth_date)
        
        entry = [
            random.choice(voornamen),
            random.choice(tussenvoegsels),
            random.choice(achternamen),
            birth_date.strftime('%Y-%m-%d'),
            random.choice(plaatsen),
            death_date.strftime('%Y-%m-%d'),
            random.choice(plaatsen),
            str(random.choice([True, False])).lower()
        ]
        
        f.write(','.join(entry) + '\n')

print("Generated sample_bidprentjes.csv with 1000 entries") 