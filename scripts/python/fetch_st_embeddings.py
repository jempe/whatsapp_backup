import sys
import json

from sentence_transformers import SentenceTransformer


model = SentenceTransformer("all-MiniLM-L6-v2")


input_text = ""

for line in sys.stdin:
    input_text += line


try:
    sentences = json.loads(input_text)

    try:
        embeddings = model.encode(sentences)

        result = {
            "embeddings": embeddings.tolist()
        }

        print(json.dumps(result))
    except Exception as e:
        error = {
            "error": "Error during embedding",
            "message": str(e)
        }
        print(json.dumps(error))
except json.JSONDecodeError as e:
    error = {
        "error": "Invalid JSON input",
        "message": str(e)
    }
    print(json.dumps(error))

