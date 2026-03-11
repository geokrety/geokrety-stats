---
title: Public Schema
description: Shared geospatial support objects used by GeoKrety and stats.
icon: material/earth
---

# `public` schema

The `public` schema is mostly infrastructure: PostGIS metadata views, spatial reference tables, and raster support used by GeoKrety geospatial features and waypoint processing.

## Inventory

- `countries`: geometry-backed country boundaries used for country derivation and geospatial joins
- `spatial_ref_sys`: PostGIS coordinate reference system catalog
- `srtm`: raster store for elevation support
- `timezones`: timezone polygons
- metadata views: `geography_columns`, `geometry_columns`, `pg_all_foreign_keys`, `raster_columns`, `raster_overviews`, `srtm_metadata`, `tap_funky`

## Why it matters to stats

- `gk_moves.position` and waypoint tables depend on PostGIS types and functions exposed through shared extensions
- country and spatial normalization in `geokrety` would not work without the underlying geospatial stack being healthy
- future map-heavy stats products will likely read country geometry from `public.countries`

## Operational notes

- `public.countries` is large enough to matter for geospatial plans at roughly `146 MB`
- `public.srtm` supports elevation-related behavior and should be treated as support data, not analytics history
- metadata views are system support and should not be exposed directly through a public API

## Operator check

- verify PostGIS and raster support are installed before diagnosing geospatial stats issues
- confirm `public.countries` and `public.srtm` are present in environments that need geospatial enrichment

## See also

- [Schema hub](specs.md)
- [Geokrety schema](specs.geokrety.md)
