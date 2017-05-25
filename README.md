# insomniac

insomniac is a GitHub review bot that checks pull requests for any
calls to `sleep` followed by a hardcoded number, and rejects all
pull requests that do this awful thing.

## running insomniac

You can run insomniac on heroku or dokku easily by just creating an
app and pushing a clone of this repo to it. Then just [follow the
instructions](https://developer.github.com/guides/building-a-ci-server/)
to set up a webhook for your repo pointing at your running instance
of insomniac. The only action that insomniac cares about is
`pull_request`, so you can give it just pull-request events. Make
sure to fill out the "Secret" with a random 20-digit-long (or
longer) string. GitHub recommends using the output of `ruby
-rsecurerandom -e 'puts SecureRandom.hex(20)'` and that seems
perfectly reasonable to me. Keep the secret handy for a minute;
we're going to need to tell insomniac about it, too.

You'll also need to [generate a personal access token for
insomniac](https://github.com/settings/tokens); it needs the full
`repo` permission if you want it to be able to review commits on
your private repositories, but you could also grant it just
`repo:status` and `public_repo` if you only want to use it to vet
public code.

You'll need to configure two environment variables for insomniac to
work properly.

- `GITHUB\_SECRET` is used to authenticate the webhook payloads GitHub
  sends
- `GITHUB\_ACCESS\_TOKEN` is a personal access token insomniac uses to
  grab diffs and leave Status feedback (pending, failure, or success)

You can set variables in Heroku with:

    [insomniac] % export SECRET=$(ruby -rsecurerandom -e 'puts SecureRandom.hex(20)')
    [insomniac] % echo $SECRET | pbcopy # go paste this in the webhook config on GitHub
    [insomniac] % heroku config:set GITHUB_SECRET=${SECRET}
    [insomniac] % heroku config:set GITHUB_ACCESS_TOKEN=<github personal access token>

Or, if you're using dokku:

    [insomniac] % export SECRET=$(ruby -rsecurerandom -e 'puts SecureRandom.hex(20)')
    [insomniac] % echo $SECRET | pbcopy # go paste this in the webhook config on GitHub
    [insomniac] % dokku config:set insomniac GITHUB_SECRET=74484095bec540d2e0c4ea42232acbb9fb357e50`
    [insomniac] % dokku config:set insomniac GITHUB_ACCESS_TOKEN=<github personal access token>
