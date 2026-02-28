<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { getCountryFlag } from '../composables/useCountryFlags.js'
import { fetchList } from '../composables/useApi.js'
import LineChart from '../components/LineChart.vue'

const route = useRoute()
const country = ref(route.params.country?.toUpperCase())
const countryData = ref(null)
const evolution = ref([])
const loading = ref(false)
const error = ref(null)

const countryNames = {
  "AF": "Afghanistan", "AX": "Åland Islands", "AL": "Albania", "DZ": "Algeria", "AS": "American Samoa", "AD": "Andorra", "AO": "Angola", "AI": "Anguilla", "AQ": "Antarctica", "AG": "Antigua and Barbuda", "AR": "Argentina", "AM": "Armenia", "AW": "Aruba", "AU": "Australia", "AT": "Austria", "AZ": "Azerbaijan",
  "BS": "Bahamas", "BH": "Bahrain", "BD": "Bangladesh", "BB": "Barbados", "BY": "Belarus", "BE": "Belgium", "BZ": "Belize", "BJ": "Benin", "BM": "Bermuda", "BT": "Bhutan", "BO": "Bolivia", "BQ": "Bonaire, Sint Eustatius and Saba", "BA": "Bosnia and Herzegovina", "BW": "Botswana", "BV": "Bouvet Island", "BR": "Brazil", "IO": "British Indian Ocean Territory", "BN": "Brunei Darussalam", "BG": "Bulgaria", "BF": "Burkina Faso", "BI": "Burundi",
  "KH": "Cambodia", "CM": "Cameroon", "CA": "Canada", "CV": "Cape Verde", "KY": "Cayman Islands", "CF": "Central African Republic", "TD": "Chad", "CL": "Chile", "CN": "China", "CX": "Christmas Island", "CC": "Cocos (Keeling) Islands", "CO": "Colombia", "KM": "Comoros", "CG": "Congo", "CD": "Congo, The Democratic Republic of the", "CK": "Cook Islands", "CR": "Costa Rica", "CI": "Cote d'Ivoire", "HR": "Croatia", "CU": "Cuba", "CW": "Curaçao", "CY": "Cyprus", "CZ": "Czech Republic",
  "DK": "Denmark", "DJ": "Djibouti", "DM": "Dominica", "DO": "Dominican Republic",
  "EC": "Ecuador", "EG": "Egypt", "SV": "El Salvador", "GQ": "Equatorial Guinea", "ER": "Eritrea", "EE": "Estonia", "ET": "Ethiopia",
  "FK": "Falkland Islands (Malvinas)", "FO": "Faroe Islands", "FJ": "Fiji", "FI": "Finland", "FR": "France", "GF": "French Guiana", "PF": "French Polynesia", "TF": "French Southern Territories",
  "GA": "Gabon", "GM": "Gambia", "GE": "Georgia", "DE": "Germany", "GH": "Ghana", "GI": "Gibraltar", "GR": "Greece", "GL": "Greenland", "GD": "Grenada", "GP": "Guadeloupe", "GU": "Guam", "GT": "Guatemala", "GG": "Guernsey", "GN": "Guinea", "GW": "Guinea-Bissau", "GY": "Guyana",
  "HT": "Haiti", "HM": "Heard Island and McDonald Islands", "VA": "Holy See (Vatican City State)", "HN": "Honduras", "HK": "Hong Kong", "HU": "Hungary",
  "IS": "Iceland", "IN": "India", "ID": "Indonesia", "IR": "Iran, Islamic Republic of", "IQ": "Iraq", "IE": "Ireland", "IM": "Isle of Man", "IL": "Israel", "IT": "Italy",
  "JM": "Jamaica", "JP": "Japan", "JE": "Jersey", "JO": "Jordan",
  "KZ": "Kazakhstan", "KE": "Kenya", "KI": "Kiribati", "KP": "Korea, Democratic People's Republic of", "KR": "Korea, Republic of", "KW": "Kuwait", "KG": "Kyrgyzstan",
  "LA": "Lao People's Democratic Republic", "LV": "Latvia", "LB": "Lebanon", "LS": "Lesotho", "LR": "Liberia", "LY": "Libya", "LI": "Liechtenstein", "LT": "Lithuania", "LU": "Luxembourg",
  "MO": "Macao", "MK": "Macedonia, The Former Yugoslav Republic of", "MG": "Madagascar", "MW": "Malawi", "MY": "Malaysia", "MV": "Maldives", "ML": "Mali", "MT": "Malta", "MH": "Marshall Islands", "MQ": "Martinique", "MR": "Mauritania", "MU": "Mauritius", "YT": "Mayotte", "MX": "Mexico", "FM": "Micronesia, Federated States of", "MD": "Moldova, Republic of", "MC": "Monaco", "MN": "Mongolia", "ME": "Montenegro", "MS": "Montserrat", "MA": "Morocco", "MZ": "Mozambique", "MM": "Myanmar",
  "NA": "Namibia", "NR": "Nauru", "NP": "Nepal", "NL": "Netherlands", "NC": "New Caledonia", "NZ": "New Zealand", "NI": "Nicaragua", "NE": "Niger", "NG": "Nigeria", "NU": "Niue", "NF": "Norfolk Island", "MP": "Northern Mariana Islands", "NO": "Norway",
  "OM": "Oman",
  "PK": "Pakistan", "PW": "Palau", "PS": "Palestine, State of", "PA": "Panama", "PG": "Papua New Guinea", "PY": "Paraguay", "PE": "Peru", "PH": "Philippines", "PN": "Pitcairn", "PL": "Poland", "PT": "Portugal", "PR": "Puerto Rico",
  "QA": "Qatar",
  "RE": "Reunion", "RO": "Romania", "RU": "Russian Federation", "RW": "Rwanda",
  "BL": "Saint Barthélemy", "SH": "Saint Helena, Ascension and Tristan da Cunha", "KN": "Saint Kitts and Nevis", "LC": "Saint Lucia", "MF": "Saint Martin (French part)", "PM": "Saint Pierre and Miquelon", "VC": "Saint Vincent and the Grenadines", "WS": "Samoa", "SM": "San Marino", "ST": "Sao Tome and Principe", "SA": "Saudi Arabia", "SN": "Senegal", "RS": "Serbia", "SC": "Seychelles", "SL": "Sierra Leone", "SG": "Singapore", "SX": "Sint Maarten (Dutch part)", "SK": "Slovakia", "SI": "Slovenia", "SB": "Solomon Islands", "SO": "Somalia", "ZA": "South Africa", "GS": "South Georgia and the South Sandwich Islands", "SS": "South Sudan", "ES": "Spain", "LK": "Sri Lanka", "SD": "Sudan", "SR": "Suriname", "SJ": "Svalbard and Jan Mayen", "SZ": "Swaziland", "SE": "Sweden", "CH": "Switzerland", "SY": "Syrian Arab Republic",
  "TW": "Taiwan, Province of China", "TJ": "Tajikistan", "TZ": "Tanzania, United Republic of", "TH": "Thailand", "TL": "Timor-Leste", "TG": "Togo", "TK": "Tokelay", "TO": "Tonga", "TT": "Trinidad and Tobago", "TN": "Tunisia", "TR": "Turkey", "TM": "Turkmenistan", "TC": "Turks and Caicos Islands", "TV": "Tuvalu",
  "UG": "Uganda", "UA": "Ukraine", "AE": "United Arab Emirates", "GB": "United Kingdom", "US": "United States", "UM": "United States Minor Outlying Islands", "UY": "Uruguay", "UZ": "Uzbekistan",
  "VU": "Vanuatu", "VE": "Venezuela, Bolivarian Republic of", "VN": "Viet Nam", "VG": "Virgin Islands, British", "VI": "Virgin Islands, U.S.",
  "WF": "Wallis and Futuna", "EH": "Western Sahara",
  "YE": "Yemen",
  "ZM": "Zambia", "ZW": "Zimbabwe"
}

const getFullCountryName = (iso) => countryNames[iso?.toUpperCase()] || iso

const formatInt = (num) => {
  if (!num) return '0'
  return Math.round(num).toLocaleString()
}

const formatFloat = (num, decimals = 2) => {
  if (!num) return '0'
  return (Math.round(num * Math.pow(10, decimals)) / Math.pow(10, decimals)).toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  })
}

async function loadCountryData() {
  loading.value = true
  error.value = null
  try {
    const response = await fetch(`/api/v1/stats/countries`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const { data } = await response.json()

    // Find this country in the data
    const found = data?.find(c => c.country.toUpperCase() === country.value)
    if (found) {
      countryData.value = found
    } else {
      error.value = `Country ${country.value} not found`
    }

    // Load evolution
    const ev = await fetchList(`/stats/countries/${country.value}/evolution/move-types`)
    evolution.value = ev.items
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadCountryData)
</script>

<template>
  <div>
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-2">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item"><RouterLink to="/countries">Countries</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">{{ country }}</li>
      </ol>
    </nav>

    <!-- Loading / Error / Content -->
    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border"></div>
    </div>
    <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-else-if="!countryData" class="alert alert-info">Country data not available.</div>
    <div v-else>
      <!-- Header -->
      <div class="card mb-4 shadow-sm">
        <div class="card-body">
          <div class="d-flex align-items-center gap-3">
            <span class="fs-1">{{ getCountryFlag(country) }}</span>
            <div>
              <h1 class="mb-1">{{ getFullCountryName(country) }}</h1>
              <p class="text-muted mb-0">Code: {{ country }} — Country activity and statistics</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Key Statistics -->
      <div class="row g-3 mb-2">
        <div class="col-12 col-md-6 col-lg-4">
          <div class="card shadow-sm border-0">
            <div class="card-body">
              <div class="text-muted small mb-2" title="Total points earned by all GeoKrety that visited this country">Total Points</div>
              <div class="fs-3 fw-bold text-success">{{ formatInt(countryData.total_points_awarded) }}</div>
              <div class="text-muted small mt-2">from {{ formatInt(countryData.total_moves) }} moves</div>
            </div>
          </div>
        </div>

        <div class="col-12 col-md-6 col-lg-4">
          <div class="card shadow-sm border-0">
            <div class="card-body">
              <div class="text-muted small mb-2" title="Average points earned per move in this country">Avg Points per Move</div>
              <div class="fs-3 fw-bold text-info">{{ formatFloat(countryData.avg_points_per_move, 4) }}</div>
              <div class="text-muted small mt-2">based on {{ formatInt(countryData.total_moves) }} total moves</div>
            </div>
          </div>
        </div>

        <div class="col-12 col-md-6 col-lg-4">
          <div class="card shadow-sm border-0">
            <div class="card-body">
              <div class="text-muted small mb-2" title="Total number of distinct users who interacted with GeoKrety in this country">Active Participants</div>
              <div class="fs-3 fw-bold text-primary">{{ formatInt(countryData.unique_users) }}</div>
              <div class="text-muted small mt-2">{{ formatInt(countryData.unique_gks) }} unique GeoKrety involved</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Country Evolution Charts -->
      <div class="row g-4 mb-2">
        <div class="col-12">
          <div class="card shadow-sm h-100">
            <div class="card-header border-0 bg-transparent py-3">
              <div class="d-flex justify-content-between align-items-center">
                <b>Move Types Evolution</b>
                <span class="badge bg-light text-dark border">Stacked Area Chart</span>
              </div>
            </div>
            <div class="card-body pt-0">
              <template v-if="evolution.length">
                <LineChart
                  :data="evolution"
                  x-key="month"
                  stacked
                  :datasets="[
                    { key: 'drops', label: 'Drops', color: '#0d6efd' },
                    { key: 'grabs', label: 'Grabs', color: '#198754' },
                    { key: 'dips', label: 'DIPs', color: '#ffc107' },
                    { key: 'seen', label: 'Seen', color: '#6c757d' },
                    { key: 'comments', label: 'Comments', color: '#f06292' }
                  ]"
                  :height="320"
                />
                <p class="text-muted small mt-2">
                  Monthly evolution of move types. Stacked areas show the relative volume of each activity.
                </p>
              </template>
              <p v-else class="text-muted text-center py-3">No activity data for this country.</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Move Type Breakdown -->
      <div class="card shadow-sm mb-2">
        <div class="card-header bg-light">
          <h5 class="mb-0">Move Type Breakdown</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
              <div class="fs-2 mb-2">🌳</div>
              <div class="text-muted small" title="GeoKrety placed into a cache">Drops</div>
              <div class="fs-5 fw-bold">{{ formatInt(countryData.drops) }}</div>
            </div>
            <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
              <div class="fs-2 mb-2">🚀</div>
              <div class="text-muted small" title="GeoKrety taken from a cache or person">Grabs</div>
              <div class="fs-5 fw-bold">{{ formatInt(countryData.grabs) }}</div>
            </div>
            <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
              <div class="fs-2 mb-2">🥾</div>
              <div class="text-muted small" title="Virtual carry - GeoKrety held digitally without physical cache">DIPs</div>
              <div class="fs-5 fw-bold">{{ formatInt(countryData.dips) }}</div>
            </div>
            <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
              <div class="fs-2 mb-2">👀</div>
              <div class="text-muted small" title="GeoKrety spotted but not taken or placed">Seen</div>
              <div class="fs-5 fw-bold">{{ formatInt(countryData.seen) }}</div>
            </div>
            <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
              <div class="fs-2 mb-2">📝</div>
              <div class="text-muted small" title="Log a comment about GeoKrety status">Comments</div>
              <div class="fs-5 fw-bold">{{ formatInt(countryData.comments) }}</div>
            </div>
            <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
              <div class="fs-2 mb-2">❤️</div>
              <div class="text-muted small" title="Favorite/love marks given to GeoKrety">Loves</div>
              <div class="fs-5 fw-bold text-danger">{{ formatInt(countryData.total_loves) }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Additional Stats -->
      <div class="card shadow-sm">
        <div class="card-header bg-light">
          <h5 class="mb-0">Summary</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <div class="mb-2">
                <div class="text-muted small" title="Count of distinct GeoKrety that visited this country or were born in this country">Unique GeoKrety</div>
                <div class="fs-5 fw-bold">{{ formatInt(countryData.unique_gks) }}</div>
              </div>
            </div>
            <div class="col-md-6">
              <div class="mb-2">
                <div class="text-muted small" title="Count of distinct users who made moves with GeoKrety in this country">Unique Users</div>
                <div class="fs-5 fw-bold">{{ formatInt(countryData.unique_users) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
a {
  color: inherit;
  text-decoration: none;
}

a:hover {
  color: #0d6efd;
}
</style>
