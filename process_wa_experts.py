import psycopg2
import datetime  # Assuming you need to generate 'date'
import re
import os

# Replace placeholders with your actual database credentials
DATABASE_CONFIG = {
    'database': 'whatsapp_backup',
    'user': 'kastro',
    'password': '',
    'host': 'localhost',
    'port': 5432
}

BRACKET_START = '___bracket_start___'

def find_string_enclosed_in(string, start, end):
    try:
        start_index = string.index(start) + len(start)
        end_index = string.index(end, start_index)
        return string[start_index:end_index]
    except ValueError:
        return ''

def convert_date_string_to_datetime(date_string):
    from datetime import datetime

    # Split the date and time parts
    date_part, time_part = date_string.split(",")

    # Define the format for the date
    date_format = "%d/%m/%y"

    # Parse the date part
    date_object = datetime.strptime(date_part.strip(), date_format)

    # Define the format for the time (assuming 24-hour format for consistency)
    time_format = "%H:%M:%S"

    # Parse the time part, handling potential leading whitespace with strip()
    time_object = datetime.strptime(time_part.strip().split(" ")[0], time_format)

    # Combine the date and time objects
    datetime_object = datetime.combine(date_object, time_object.time())

    return datetime_object


def save_message_to_database(message_data):
    date = message_data['date'] 
    phone_number = message_data['phone_number']
    message = message_data['message']
    attachment = message_data['attachment']

    try:
        # Connect to the database
        with psycopg2.connect(**DATABASE_CONFIG) as conn:
            with conn.cursor() as cur:
                # SQL query with placeholders to prevent SQL injection
                query = """
                    INSERT INTO messages (message_date, phone_number, message, attachment)
                    VALUES (%s, %s, %s, %s)
                """
                cur.execute(query, (date, phone_number, message, attachment))
    except psycopg2.Error as e:
        print("Database error:", e)


filename = '/Users/kastro/Downloads/WhatsApp Chat - UP/_chat.txt'

# Read the file
with open(filename, 'r') as file:
    data_file = file.read()

date_regex =  r"^\[\d+/\d+/\d+, \d+:\d+:\d+\u202f(A|P)M\]"

start_bracket_regex = r"^\["
start_space_regex = r"^\u200e"

data = ''

lines = data_file.split('\n')

for line in lines:
    if (re.match(start_bracket_regex, line) and not re.match(date_regex, line)):
        line = re.sub(start_bracket_regex, BRACKET_START, line)

    if (re.match(start_space_regex, line)):
        line = re.sub(start_space_regex, '', line)

    data += line + '\n'

#replace the first character of the file
data = data.replace('[', '', 1)

# Split the data into lines
lines = data.split('\n[')

messages = []


for line in lines:

    date_parts = line.split(']')

    date_string = date_parts[0].replace('[', '')

    message = date_parts[1].replace(BRACKET_START, '[').strip()
   
    attachment = find_string_enclosed_in(line, '<attached:', '>')
    phone_number = find_string_enclosed_in(line, '\u202a', '\u202c')

    message = message.replace('<attached:' + attachment + '>', '');
    message = message.replace('\u202a' + phone_number + '\u202c:', '');

    date = convert_date_string_to_datetime(date_string)

    formatted_date = date.strftime('%Y-%m-%d %H:%M:%S')

    message_data = {
        'date': formatted_date,
        'phone_number': find_string_enclosed_in(line, '\u202a', '\u202c'),
        'message': message,
        'attachment': attachment.strip()
    }

    save_message_to_database(message_data)


    #TODO: Fix issue with messages that contains [

    #print(f"message:  {line}")

    print(message_data)

    #if line.find('\u200e') != -1 and line.find('\u202a') == -1:
    #    print(line)


