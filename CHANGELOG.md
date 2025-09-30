# Changelog

## v0.5.0 (30 Sep 2025)

- Added "verify" command to validate campaigns and recipient lists
- Added JSON Schema validation for recipient parameters
- Extended DKIM private key encoding support
- Security update for go-mail dependency
- Upgraded UI to Tailwind CSS v4
- Enabled golangci-lint

## v0.4.0 (22 Sep 2025)

- Migrated to [wneessen/go-mail](github.com/wneessen/go-mail)
- Fixed frontmatter error handling @guyou
- Allow custom "To" template @guyou
- Added attachment support @guyou

## v0.4.0-beta.2 (20 Sep 2025)

- Switched to Cobra's Context
- Enhanced TLS configuration @gbonnefille
- Upgraded dependencies & added tests
- Upgraded to Ember 6.7.0

## v0.4.0-beta.1 (28 Jun 2025)

- Added sendCampaign with ZipFs
- Added legacy TLS support @gbonnefille
- Added API middleware for panics, logging, and auth
- Support for `.yml` file extension @gbonnefille
- Upgraded to Go 1.24 & Ember 6.5
- Refactored `cmd` code structure

## v0.3.0-beta.1 (24 Jun 2024)

- Minor fixes & dependency updates
- Upgraded to Ember 5.9
- Upgraded to Go 1.22

## v0.3.0-alpha.1 (18 Aug 2021)

- Added Ember-based UI for preview

## 0.2.1 (16 Jul 2021)

- Reset connection on failure
- Render text emails with Glamour
- Use multiple workers for delivery
- Configurable # of workers and send rate
- Switched dependecy management to Go Modules
- Switched to Goldmark for Markdown rendering
- Campaign subject is now also a template
- Removed usage of global configurations
- Preview UI proxying & config
- Added sendBeta API

## 0.2.0 (25 Jul 2017)

- Added GraphQL API for listing and rendering campaigns
- Added "server" command to run GraphQL API
- Added "preview" command to open browser preview

## 0.1.0 (11 Jul 2017)

- Initial preview release
- Basic project initialization functionality 
- Rendering campaign Markdown with CSS inlining
- Delivery of campaigns via SMTP
- DKIM signing
