- You MUST fix all the items below
- as soon as a task is done you MUST instantanely check it in this #file:docs/api/fixes.md
- remove the "TODO" comment related to the just accomplished task. Keep the one unrelated to the change
-

## in #file:api/internal/api/router.go

- [ ] implement route /geokrety/ in #file:api/internal/api/router.go
- [ ] rename `statsHandler.GetGeokretyById` to `statsHandler.GetGeokretyDetailsById`
- [ ] all `/geokrety/{id}/*` must redirect to the `gkid` version `/geokrety/{gkid}`. Ex:
  - `GET /geokrety/{id}` -> REDIRECT `GET /geokrety/{gkid}`
  - `GET /geokrety/{id}/loved-by` -> REDIRECT `GET /geokrety/{gkid}/loved-by`
  - etc...
- [ ] implement statsHandler.GetUserList
- [ ] implement statsHandler.GetPictureList
- [ ] rename `statsHandler.GetPicture` to `statsHandler.GetPictureDetails`
- [ ] implement statsHandler.GetCountryList and add `cr.Get("/", statsHandler.GetCountryList)`
- [ ] implement `statsHandler.GetGeokretyDetailsByGkId`
  - [ ] lookup by GKID
