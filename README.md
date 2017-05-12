# insomniac

insomniac is a GitHub review bot that checks pull requests for any
calls to `sleep` followed by a hardcoded number, and rejects all
pull requests that do this awful thing.

## running insomniac

You can run insomniac on heroku or dokku easily by just creating an
app and pushing a clone of this repo to it. Then just [follow the
instructions](https://developer.github.com/guides/building-a-ci-server/)
to set up a webhook for your repo pointing at your running instance
of insomniac.
