
# GeoKrety Stats Project

## Repo organization

The repository contains the [API](./api/) and the [frontend][./frontend].

## GeoKrety IMPORTANT info

The table `geokrety.gk_geokrety` contains columns `id` and `gkid`, the `id` is only used *internally by database for foreign keys*.

The `gkid` (in its string version eg: `GK0001`, `GK3D45F`) is the version the *user must see and use*. All API endpoint related to `GeoKrety` must accept "integer `gkid`" and "string `gkid`".

## GeoKrety concepts

### convert integer `gkid` <=> string `gkid`

To convert between integer GeoKret IDs and the public GKID format. Mapping rule (from geokrety.org):

```
GKID = "GK" + integer.toString(16).toUpperCase().padStart(4, "0")
```

Examples:
```
  intToGkid(1)        → "GK0001"
  intToGkid(255)      → "GK00FF"
  intToGkid(65535)    → "GKFFFF"
  gkidToInt("GK00FF") → 255
```
