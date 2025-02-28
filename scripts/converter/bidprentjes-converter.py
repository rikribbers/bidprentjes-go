import pylightxl as xl
from datetime import datetime


# read only selective sheetnames
db = xl.readxl(fn='bidprentjes.xlsx', ws=('website'))

# extract a datetime Object or Nono


def extractDate(dateStr):
    try:
        return datetime.strptime(dateStr, '%Y/%m/%d')
    except:
        try:
            return datetime.strptime(dateStr, '%d-%m-%Y')
        except:
            try:
                return datetime.strptime(dateStr[:-9], '%d-%m-%Y')
            except:
                if dateStr == '':
                    return None
                else:
                    print('Parsing date failed for id: ' +
                          str(id) + ' dataStr: ' + dateStr)
                    return None


with open('bidprentjes.csv', 'w') as output_file:

    i = 0
    for row in db.ws(ws='website').rows:
        # id,geboren,overleden,achternaam,geboorteplaats,voorvoegsel,voornaam,rustplaats,scan

        id = row[0]
        result = str(id) + ","

        voornaam = row[6]
        if voornaam != '':
            result = result + '"' + str(voornaam) + '"'

        result = result + ','

        voorvoegsel = row[5]
        if voorvoegsel != '':
            result = result + '"' + str(voorvoegsel) + '"'

        result = result + ','

        achternaam = row[3]
        if achternaam != '':
            result = result + '"' + str(achternaam) + '"'

        result = result + ','

        geboren = extractDate(row[1])
        geborenStr = ''
        if geboren != None:
            result = result + geboren.strftime('%Y-%m-%d')

        result = result + ','

        geboorteplaats = row[4]
        if geboorteplaats != '':
            result = result + '"' + str(geboorteplaats) + '"'

        result = result + ','

        overleden = extractDate(row[2])
        overledenStr = ''
        if overleden != None:
            result = result + overleden.strftime('%Y-%m-%d')

        result = result + ','

        rustplaats = row[7]
        if rustplaats != '':
            result = result + '"' + str(rustplaats) + '"'

        result = result + ','

        scan = row[8]
        if scan.lower() == 'ja':
            result = result + 'true\n'
        else:
            result = result + 'false\n'

        output_file.write(result)
