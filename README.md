# keysync

`keysync` is a CLI tool for keeping the `authorized_keys` file on your servers
in sync with your GitHub user's public keys.

This is done by connecting to the GitHub API and fetching the [user's publicly
available keys](https://developer.github.com/v3/users/keys/#list-public-keys-for-a-user).

The tool will keep existing keys intact whilst adding the keys from GitHub separately.
Keys removed from GitHub will also be removed from the file.

## Usage

To use the CLI tool, download the [latest version](https://github.com/dhensby/keysync/releases/latest)
for your system and get started by running `keysync -gh-user [username]` as the
user you wish to update the keys for.

The tool can be run on behalf of other users on the server `keysync -user [localuser] -gh-user [username]`.

## Setting up a cronjob

You can set up a cronjob to run as often as you like to keep things in sync. Run
`crontab -e` as your user, or `crontab -e -u [localuser]` to set one up on behalf
of another user, and add the following to sync your keys every 5 minutes:

```cron
*/5 * * * * /path/to/keysync -gh-user octocat
```

## Pushover integration

You can get simple push notifications to your phone via [Pushover](https://pushover.net/).
Create an account, and add a `.keysync.config` file in the same directory as the executable;
the file should be a JSON file formatted as so:

```json
{
    "PushoverAppKey": "...",
    "PushoverUserKey": "..."
}
```