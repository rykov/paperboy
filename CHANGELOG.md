# Changelog

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
