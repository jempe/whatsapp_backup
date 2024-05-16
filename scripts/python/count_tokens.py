import tiktoken
import sys
import json

input_text = ""

for line in sys.stdin:
    input_text += line

def num_tokens_from_string(string: str, encoding_name: str) -> int:
    """Returns the number of tokens in a text string."""
    encoding = tiktoken.get_encoding(encoding_name)
    num_tokens = len(encoding.encode(string))
    return num_tokens

result = {
    "num_tokens": num_tokens_from_string(input_text, "cl100k_base")
}

# Output the result as a JSON string
print(json.dumps(result))

