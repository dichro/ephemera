The main binary is `twitter` in this directory. Before you can use it usefully, you must:

1. Create an application ID at https://apps.twitter.com/. Create an access token for yourself too; copy all four fields into ~/.config/ephemera.yaml as `twitter_id` (the application ID! Not your username!), `twitter_secret`, `twitter_access_token`, and `twitter_access_secret`.

2. Set `store` to point to some unobtrusive directory like `~/ephemera.db`; that will store all details from your timeline, including deleted tweets, in case you need to respond to a subpoena.

2. Download all your Twitter data from https://twitter.com/settings/your_twitter_data - this is functionally optional if you have fewer than 3200 tweets in your timeline, but don't you want a backup anyway?

3. Run `twitter timeline fetch` to retrieve the most recent 3200 tweets (API limit for scanning through a timeline).

4. Run `twitter archive <filename>` with the filename of the archive you downloaded from Twitter to load older tweets beyond the limit of 3200.

5. Experiment with `twitter timeline policy`, `twitter timeline policy drops`, and `twitter timeline policy keeps` to see what would be deleted or retained from your timeline. If you don't like the results, edit cmd/tl_policy.go until you do.

6. Run `twitter timeline sanitize` to apply the policy and delete tweets.

Still to come: sanitizing DMs, likes, and maybe even followers/lists, somehow.