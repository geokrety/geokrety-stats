---
title: Secure Schema
description: Restricted schema for sensitive key material.
icon: material/lock
---

# `secure` schema

The `secure` schema is intentionally minimal and sensitive.

## Inventory

- `gpg_keys`: public and private key material

## Documentation stance

This reference records that the schema exists and is part of the live database contract. It does not document key contents, operational secrets, or handling procedures beyond the minimum structural description.

## Operational notes

- restrict access sharply
- exclude from public read APIs
- avoid copying this schema into lower-trust environments unless required for a specific test or operational workflow

## Operator check

- document which environments are allowed to hold real key material
- verify redaction of secrets in backups, dumps, and support bundles

## See also

- [Schema hub](specs.md)
