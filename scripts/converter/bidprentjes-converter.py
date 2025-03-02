import pylightxl as xl
from datetime import datetime


# read only selective sheetnames
db = xl.readxl(fn='bidprentjes.xlsx', ws=('website'))

# extract a datetime Object or None
def extractDate(dateStr):
    if not dateStr or dateStr.strip() == '':
        return None
    
    # Clean the date string by removing trailing " 0"
    dateStr = dateStr.strip()
    if dateStr.endswith(" 0"):
        dateStr = dateStr[:-2]
    
    # If the string contains a time component (00:00:00), remove it
    if " 00:00:00" in dateStr:
        dateStr = dateStr.replace(" 00:00:00", "")
    
    try:
        return datetime.strptime(dateStr, '%Y/%m/%d')
    except:
        try:
            return datetime.strptime(dateStr, '%d-%m-%Y')
        except:
            try:
                # Handle any remaining dates with time component by truncating
                if len(dateStr) > 10:
                    dateStr = dateStr[:10]
                return datetime.strptime(dateStr, '%d-%m-%Y')
            except:
                print('Parsing date failed for dataStr:', dateStr)
                return None

def clean_field(field):
    if field == '':
        return ''
    # Convert to string and remove any nested quotes and parentheses
    field = str(field).replace('"', '').replace('(', '').replace(')', '')
    # Remove any trailing commas
    field = field.rstrip(',')
    # If the field contains a comma, wrap it in quotes
    if ',' in field:
        return f'"{field}"'
    return field

with open('bidprentjes.csv', 'w') as output_file:
    i = 0
    for row in db.ws(ws='website').rows:
        # id,geboren,overleden,achternaam,geboorteplaats,voorvoegsel,voornaam,rustplaats,scan

        id = row[0]
        result = str(id) + ","

        voornaam = row[6]
        result = result + clean_field(voornaam)

        result = result + ','

        voorvoegsel = row[5]
        result = result + clean_field(voorvoegsel)

        result = result + ','

        achternaam = row[3]
        result = result + clean_field(achternaam)

        result = result + ','

        geboren = extractDate(row[1])
        result = result + ','  # Always add comma for empty date
        if geboren is not None:
            result = result[:-1] + geboren.strftime('%Y-%m-%d') + ','  # Replace last comma with formatted date

        geboorteplaats = row[4]
        result = result + clean_field(geboorteplaats)

        result = result + ','  # Always add comma for empty date
        overleden = extractDate(row[2])
        if overleden is not None:
            result = result[:-1] + overleden.strftime('%Y-%m-%d') + ','  # Replace last comma with formatted date

        result = result + ','

        rustplaats = row[7]
        result = result + clean_field(rustplaats)

        result = result + ','

        scan = row[8]
        if scan.lower() == 'ja':
            result = result + 'true\n'
        else:
            result = result + 'false\n'

        output_file.write(result)
