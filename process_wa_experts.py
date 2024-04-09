import psycopg2
import datetime  # Assuming you need to generate 'date'

# Replace placeholders with your actual database credentials
DATABASE_CONFIG = {
    'database': 'whatsapp_backup',
    'user': 'kastro',
    'password': '',
    'host': 'localhost',
    'port': 5432
}

def find_string_enclosed_in(string, start, end):
    try:
        start_index = string.index(start) + len(start)
        end_index = string.index(end, start_index)
        return string[start_index:end_index]
    except ValueError:
        return ''

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
    data = file.read()

#replace the first character of the file
data = data.replace('[', '', 1)

# Split the data into lines
lines = data.split('\n[')

messages = []


for line in lines:

    date_parts = line.split(']')

    date = date_parts[0].replace('[', '')

    message = date_parts[1].strip()
   
    attachment = find_string_enclosed_in(line, '<attached:', '>')
    phone_number = find_string_enclosed_in(line, '\u202a', '\u202c')

    message = message.replace('<attached:' + attachment + '>', '');
    message = message.replace('\u202a' + phone_number + '\u202c:', '');

    message_data = {
        'date': date,
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


