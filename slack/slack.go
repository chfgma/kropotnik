package slack

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const response = `Thank you for stepping up and taking on this delivery! Here are the full details: %s

Please follow up directly with this individual via their preferred form of contact to arrange a good time to drop off and confirm any further details - e.g. a shopping list.

Shopping guidance:
If this person can cover the cost of their groceries themselves, that’s great - have them reimburse you directly. If not, we are currently covering up to $50 for one person, or $100 for households with 2 or more once per week.   Please take a picture of your receipt and submit to #community_fund_mgmt and you will be reimbursed within 24 hours. Please try to buy about a week’s worth of food for the household.  It’s ok if you can’t find everything on the list, we know a lot is stocked out; just try to meet nutritional needs.

Please be sure to be safe and sanitary in delivering:
- wash your hands before, after
- maintain 6 feet of distance at all times
- leave the grocery bag at the door and ring the bell, knock or call to let them know it’s arrived
- remind the person to wash what they receive with soap and/or disinfectant if they have it

Ask the person if they’d like a follow up check-in, and if they need any other kind of support.  If you’re willing to do this check, we encourage you to; it’s good to form longer-lasting connections with your neighbors, if this is something they want as well.

When the delivery is complete, please respond to the original thread in <#C010M24QT4G> (where you picked up this request) with a note to @intake that the delivery is complete, and if there is any follow up needed by you or someone else.

If you run into any challenges along the way (someone’s not answering their phone, they have more complex needs than you’re able to address), post to @intake (in the request thread or <#C010M24QT4G>) and we’ll follow up.`

func BasicsCommand(w http.ResponseWriter, r *http.Request) {
	if signingSecret, ok := os.LookupEnv("SLACK_SIGNING_SECRET"); ok {
		verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Printf("error creating slack secret verifier %v", err)

			return
		}

		if err := verifier.Ensure(); err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			log.Printf("error verifying slack request %v", err)

			return
		}
	} else {
		log.Println("not validating incoming slack request because SLACK_SIGNING_SECRET is not set")
	}

	basics, err := slack.SlashCommandParse(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		log.Printf("error parsing basics command %v", err)

		return
	}

	w.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(slackevents.MessageActionResponse{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         fmt.Sprintf(response, basics.Text),
	}); err != nil {
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		log.Printf("error writing response %v", err)

		return
	}
}
