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
