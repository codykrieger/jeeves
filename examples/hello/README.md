# hello example

These instructions will get you up and running with a simple development-only
Alexa Skill that runs on Heroku. Configuring the service for production is
beyond the scope of this document.

## prerequisites

- A working Go 1.5+ installation and configured `GOPATH` (beyond the scope of
  this document)
- A Heroku account and the Heroku Toolbelt (https://toolbelt.heroku.com)
- `godep` (`go get github.com/tools/godep`)

## set up the web service

This example service is meant to be deployed to Heroku. To do so:

```shell
# Clone this repository:
git clone git@github.com:codykrieger/jeeves.git

mkdir ask-example
# NOTE: The directory must be named ask-example, or you'll have to change the
# 'web' entry in the Procfile to match the name of the directory.

cp jeeves/examples/hello/* ask-example
cd ask-example

# Vendor the example app's dependencies:
godep save

# Initialize a new git repository and commit everything:
git init
git add -A
git commit -m "Initial commit."

# Create a new Heroku app:
heroku create
```

Take note of the URL that Heroku assigns to your app. You'll need it in a
moment.

## create a new skill

To configure your Echo to use this service, you need to:

- Register your Echo for development
- Register a new Alexa Skill and fill out a handful of required fields

See the [Testing an Alexa Skill][tas-docs] documentation for instructions on how
to do those things.

Configure the following fields like so:

- **Endpoint**: Select the `HTTPS` endpoint type, and enter your Heroku app's
  URL, suffixed with `/skills/hello`
- **Intent Schema**: Copy the contents of `intent-schema.json` into the text
  field
- **Sample Utterances**: Copy the contents of `sample-utterances.txt` into the
  text field
- **SSL Certificate:** Select the `My development endpoint is a subdomain of a
  domain that has a wildcard certificate from a certificate authority` option

The rest of the fields should be relatively self-explanatory. Once you've filled
out all of the requisite fields, enable your skill for testing under the Test
configuration tab (if it isn't already enabled).

Make note of the Application ID assigned to your skill (it should look something
like `amzn1.echo-sdk-ams.app.000000-d0ed-0000-ad00-000000d00ebe`, and is
displayed on the Skill Information configuration tab). You'll need it in a
moment.

## set your application id and deploy the app

Replace `YOUR_APPLICATION_ID_HERE` with your Alexa Skill's assigned Application
ID, and run these commands to finish deploying the app:

```shell
heroku config:set ASK_APP_ID="YOUR_APPLICATION_ID_HERE"
git push heroku
```

## test your new skill

At this point, your skill should be ready for testing!

Using the Invocation Name selected when registering your skill with Amazon, you
can say things like:

```
Alexa, launch [INVOCATION NAME].
```

And:

```
Alexa, ask [INVOCATION NAME] to say hello.
```

## next steps

To add more useful functionality to your service, have a look at the
[Alexa Skills Kit documentation][ask-docs].

[tas-docs]: https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/testing-an-alexa-skill
[ask-docs]: https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit
