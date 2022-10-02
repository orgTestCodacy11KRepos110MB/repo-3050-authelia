---
title: "Documentation"
description: "Authelia Development Documentation Guidelines"
lead: "This section covers the guidelines we use when writing documentation."
date: 2021-01-30T19:29:07+11:00
draft: false
images: []
menu:
  contributing:
    parent: "guidelines"
weight: 320
toc: true
---

## Domains

Always use the generic domain (or subdomain of) `example.com` in documentation.

If it's necessary to utilize more than one domain please ask for specific feedback in any PR.

## Certificates

When including certificates in documentation always ensure they are valid for 1 year starting at `Jan 1 00:00:00 1970`.
This ensures the certificate is not valid for multiple reasons.

In addition the guidance for [Private Keys](#private-keys) should be followed.

## Private Keys

Always append invalid data to the END of the PEM block before the base64 padding `=` (if present). The suggested
text is `^invalid DO NOT USE`. This both has an invalid base64 character `^` and has information to communicate that
users should not use the PEM block.