# kropotnik
Helper bot for Clinton Hill / Fort Greene Mutual Aid

Right now it:
- Receives calls from Twilio and writes the caller's responses to an Airtable table
- Handles the `/basics` command in the CHFG Slack

## Usage
Required environment variables:
```
AIRTABLE_API_KEY
AIRTABLE_BASE_ID
```

Optional:
```
SLACK_SIGNING_SECRET
PORT
```

Run directly with Go:
```
go run main.go
```

Run through Docker:
```
docker build --tag kropotnik:dev .
docker run --rm --envfile=/tmp/kropotkin.env -p 127.0.0.1:8080:8080 kropotnik:dev
```

The `-p 8080:8080` will map 8080 from inside of the container to your localhost:8080.
The `--envfile=/tmp/kropotkin.env` tells Docker to pull environment secrets
from a file at `/tmp/kropotkin.env` (you can store it wherever though).

As an example kropotkin.env:

```
AIRTABLE_API_KEY="keys8888888888888"
AIRTABLE_BASE_ID="appabcdefghijklmn"
SLACK_SIGNING_SECRET="50550505050"
```

Development credentials should be created so as not to interfere with our real servers and live numbers.

Once you have a local server running, you can use [ngrok](https://ngrok.com/)
to expose your server to the Internet and connect it to Slack / Twilio webhooks
for testing.
