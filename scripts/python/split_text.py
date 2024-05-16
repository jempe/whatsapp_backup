from langchain.text_splitter import RecursiveCharacterTextSplitter
import tiktoken
import sys
import json

if len(sys.argv) != 2:
    print("Error: Expected 1 argument, got", len(sys.argv) - 1)
    sys.exit(1)

input_text = ""

chunk_size = 1000
chunk_overlap = 0

chunk_arg = int(sys.argv[1])

if chunk_arg > 0:
    chunk_size = chunk_arg

for line in sys.stdin:
    input_text += line

splitter = RecursiveCharacterTextSplitter.from_tiktoken_encoder(
    chunk_size=chunk_size, chunk_overlap=chunk_overlap, encoding_name="cl100k_base"
)

parts = splitter.split_text(input_text)

result = {
    "parts": parts,
}

# Output the result as a JSON string
print(json.dumps(result))

