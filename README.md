# password-generator
This is a customizable password generator that creates strong and unique passwords based on a wordlist. <br> 
The generator combines three words, adds a random symbol between the words, capitalizes random letters, maps random letters, and inserts symbols randomly.

## Usage
1. Prepare a wordlist and place it in `src/backend/wordlists/`
2. Modify configuration in `src/backend/values/values.yaml`
```yaml
MIN_PASSWORD_LENGTH: 15          # minimum length of the password
MAX_PASSWORD_LENGTH: 32          # maximum length of the password
BETWEEN_SYMBOLS: ""            # define symbols for between the words
INSIDE_SYMBOLS: ""             # define symbols for the words
PASSWORD_PER_ROUTINE: 300      # generated passwords per GO routine          
SYMBOL_MAPPING:                # define which char you want to be swapped     
  key: value                   # value is mapped to key
```
3. Build and Run Docker image
```
docker build -t password-generator . 
docker run -p 8080:8080 password-generator
```