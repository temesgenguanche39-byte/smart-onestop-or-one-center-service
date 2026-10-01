// Smart One-Stop Digital Citizen Hearing & Automated Escalation Platform
// Frontend Application Controller

let currentLang = 'en';
let currentToken = '';
let currentUser = null;
let municipalTree = [];
let serviceTypes = [];
let allCases = [];
let attachedFiles = [];
let activeCountdownInterval = null;

const SESSION_TOKEN_KEY = 'smart_onestop_token';
const SESSION_USER_KEY = 'smart_onestop_user';

// ============================================================================
// DYNAMIC BACKEND GATEWAY RESOLUTION & VERCEL-TO-DOCKER ROUTING
// ============================================================================
const API_BASE_KEY = 'smart_onestop_api_base';

function getApiBase() {
  // 1. Check URL parameters: ?api=... or ?api_base=... or ?backend=...
  try {
    const urlParams = new URLSearchParams(window.location.search);
    const paramApi = urlParams.get('api') || urlParams.get('api_base') || urlParams.get('backend');
    if (paramApi) {
      let clean = paramApi.trim().replace(/\/+$/, '');
      localStorage.setItem(API_BASE_KEY, clean);
      return clean;
    }
  } catch (e) {
    console.warn("Could not parse search params:", e);
  }

  // 2. Check localStorage
  try {
    const stored = localStorage.getItem(API_BASE_KEY);
    if (stored && stored.trim() !== '') {
      return stored.trim().replace(/\/+$/, '');
    }
  } catch (e) {
    console.warn("Could not read localStorage:", e);
  }

  // 3. Check window.ENV
  if (window.ENV && window.ENV.API_BASE) {
    return window.ENV.API_BASE.trim().replace(/\/+$/, '');
  }

  // 4. Default: empty string (same-origin relative URL)
  return '';
}

function setApiBase(url) {
  if (!url || url.trim() === '') {
    localStorage.removeItem(API_BASE_KEY);
  } else {
    localStorage.setItem(API_BASE_KEY, url.trim().replace(/\/+$/, ''));
  }
}

function apiUrl(endpoint) {
  const base = getApiBase();
  if (!base) return endpoint;
  const cleanEndpoint = endpoint.startsWith('/') ? endpoint : '/' + endpoint;
  return `${base}${cleanEndpoint}`;
}

async function testApiConnection(customUrl) {
  const base = customUrl !== undefined ? customUrl.trim().replace(/\/+$/, '') : getApiBase();
  const testUrl = base ? `${base}/healthz` : '/healthz';
  const startTime = performance.now();
  try {
    const res = await fetch(testUrl, {
      method: 'GET',
      mode: 'cors',
      headers: { 'Bypass-Tunnel-Reminder': 'true' }
    });
    const latency = Math.round(performance.now() - startTime);
    if (res.ok) {
      const data = await res.json().catch(() => ({}));
      return { ok: true, latency, data, url: testUrl };
    }
    return { ok: false, status: res.status, latency, url: testUrl };
  } catch (err) {
    return { ok: false, error: err.message, url: testUrl };
  }
}

async function checkGatewayHealth() {
  const statusChip = document.getElementById('systemStatusChip');
  const statusText = document.getElementById('systemStatusText');
  const base = getApiBase();

  const result = await testApiConnection();
  if (result.ok) {
    if (statusChip) {
      statusChip.classList.remove('offline');
      statusChip.classList.add('online');
    }
    if (statusText) {
      if (base) {
        try {
          const u = new URL(base);
          statusText.textContent = `Gateway • ${u.host} (${result.latency}ms)`;
        } catch {
          statusText.textContent = `Gateway Connected (${result.latency}ms)`;
        }
      } else {
        statusText.textContent = currentLang === 'am' ? 'የአዲስ አበባ ማዘጋጃ ቤት ክላውድ • ንቁ' : 'Addis Ababa Municipal Cloud • Active';
      }
    }
  } else {
    if (statusChip) {
      statusChip.classList.add('offline');
      statusChip.classList.remove('online');
    }
    if (statusText) {
      statusText.textContent = currentLang === 'am' ? 'ጌትዌይ ተቋርጧል • ለማስተካከል ይጫኑ' : 'Gateway Offline • Click to Connect';
    }
  }
  return result;
}

function openModal(modalId) {
  const modal = document.getElementById(modalId);
  if (modal) modal.classList.remove('hidden');
}

function closeModal(modalId) {
  const modal = document.getElementById(modalId);
  if (modal) modal.classList.add('hidden');
}

window.openModal = openModal;
window.closeModal = closeModal;

function openGatewayModal() {
  const modal = document.getElementById('gatewayModal');
  const input = document.getElementById('gatewayUrlInput');
  const statusBox = document.getElementById('gatewayStatusDisplay');
  if (!modal) return;

  if (input) {
    input.value = getApiBase();
  }
  if (statusBox) {
    statusBox.innerHTML = '<span style="color:var(--text-tertiary);">Testing connection to active endpoint...</span>';
  }
  modal.classList.remove('hidden');

  // Run live ping test
  pingCurrentGateway();
}

async function pingCurrentGateway() {
  const input = document.getElementById('gatewayUrlInput');
  const statusBox = document.getElementById('gatewayStatusDisplay');
  const pingBtn = document.getElementById('gatewayPingBtn');
  if (!statusBox) return;

  const targetUrl = input ? input.value : getApiBase();
  if (pingBtn) pingBtn.disabled = true;
  statusBox.innerHTML = '<span>⏳ Pinging gateway endpoint...</span>';

  const res = await testApiConnection(targetUrl);
  if (pingBtn) pingBtn.disabled = false;

  if (res.ok) {
    const env = res.data?.environment || 'docker-local';
    const srv = res.data?.service || 'smart-onestop-platform';
    statusBox.innerHTML = `
      <div style="color:var(--emerald-400); font-weight:700; display:flex; align-items:center; gap:0.5rem;">
        <span>✅ CONNECTED (${res.latency}ms)</span>
      </div>
      <div style="font-size:0.75rem; color:var(--text-secondary); margin-top:0.25rem;">
        Service: <strong>${srv}</strong> • Environment: <strong>${env}</strong> • URL: <code>${res.url}</code>
      </div>
    `;
  } else {
    statusBox.innerHTML = `
      <div style="color:#ef4444; font-weight:700; display:flex; align-items:center; gap:0.5rem;">
        <span>❌ UNREACHABLE / OFFLINE</span>
      </div>
      <div style="font-size:0.75rem; color:var(--text-secondary); margin-top:0.25rem;">
        Error: ${res.error || 'HTTP ' + res.status} • Tried: <code>${res.url}</code>
      </div>
    `;
  }
}

async function saveGatewaySettings() {
  const input = document.getElementById('gatewayUrlInput');
  if (!input) return;
  const newUrl = input.value.trim();
  setApiBase(newUrl);

  showToast('Gateway configuration saved! Testing & refreshing...', 'info');
  await checkGatewayHealth();

  // Reload foundational data
  try {
    await loadMunicipalHierarchy();
    await loadServiceTypes();
    if (currentToken) {
      await refreshOfficialDashboard();
    }
    showToast('Municipal records synchronized successfully!', 'success');
  } catch (err) {
    showToast('Failed to sync data: ' + err.message, 'warning');
  }

  closeModal('gatewayModal');
}

function setGatewayPreset(presetUrl) {
  const input = document.getElementById('gatewayUrlInput');
  if (input) {
    input.value = presetUrl;
    pingCurrentGateway();
  }
}


// Bilingual Localization Dictionary
const i18n = {
  en: {
    citizenPortal: "Citizen Portal",
    officialDesk: "Official Command",
    activeServiceStatus: "Municipal Digital Gateway • Addis Ababa",
    heroTitle: "Transparent Public Service & Automated SLA Resolution",
    heroDesc: "Eliminate in-person bureaucracy. Submit grievances directly to Woreda officers, track deterministic SLA timers, and participate in Wednesday & Friday virtual hearings.",
    submitGrievance: "Submit Grievance",
    trackTicket: "Track Ticket & QR",
    activeCasesCount: "Active Grievances",
    breachedCount: "Auto-Escalating",
    hearingDays: "Digital Hearing Days",
    newComplaint: "File New Grievance",
    trackExisting: "Track Status & Hearing Slot",
    intakeFormTitle: "Citizen Grievance Submission",
    intakeFormSubtitle: "Complete the form below to receive your cryptographic ticket and SLA countdown.",
    fullName: "Full Name *",
    phoneNumber: "Phone Number (SLA SMS Alerts) *",
    nationalId: "National ID (Fayda / Kebele)",
    houseNumber: "House Number / Kebele",
    subCity: "Sub-City Administration *",
    woreda: "Woreda Jurisdiction (Initial Intake Tier) *",
    serviceType: "Service Classification & Mandatory SLA Target *",
    priority: "Urgency Priority",
    grievanceTitle: "Grievance Subject / Title *",
    detailedDescription: "Detailed Description & Previous Office Responses *",
    evidenceUpload: "Supporting Evidence (Contracts, Denials, Audio Note)",
    submitGrievanceBtn: "Submit Grievance & Generate Ticket",
    trackBtn: "Track Grievance",
    demoTickets: "Quick Demo Tickets:",
    triggerSla: "Run SLA Escalation Engine",
    kpiTotal: "Total Ingested Grievances",
    kpiActive: "Active In-Flight Cases",
    kpiBreached: "Breached SLA Thresholds",
    kpiEscalated: "Tier Escalations (Sub-City/City)",
    kpiResolved: "Certified Resolutions",
    grievanceQueue: "Case Review Queue",
    hearingDesk: "Wed & Fri Digital Hearings",
    bottleneckHeatmap: "Administrative Bottleneck Heatmap",
    auditLedger: "Immutable Audit Ledger",
    hearingTitle: "Digital Hearing Desk (Wednesdays & Fridays)",
    hearingSubtitle: "In-app virtual hearing room scheduler replacing in-person office queues per municipal protocol.",
    selectHearingDay: "Select Presiding Hearing Session:",
    voiceStudioTitle: "Citizen Voice Memo Recording (የድምጽ መልእክት መቅረጫ)",
    voiceStudioDesc: "Explain your grievance verbally in Amharic or English. Transcribed automatically into OCR text.",
    startRecording: "Record Voice Memo",
    stopRecording: "Stop & Save Memo",
    step1Title: "1. Citizen Profile",
    step2Title: "2. Jurisdiction & SLA",
    step3Title: "3. Grievance & Proof",
    nextStep: "Next Step →",
    prevStep: "← Previous Step",
    systemOnline: "Addis Ababa Municipal Cloud • Active",
    staffManagement: "Staff & RBAC Accounts",
    loginTitle: "Municipal Staff Authentication",
    loginSubtitle: "Official Command Center & Digital Hearing Desk",
    loginDesc: "Authorized personnel only. Access hearing schedules, manage escalated grievances, and sign immutable resolutions.",
    emailAddress: "Official Email Address *",
    password: "Password *",
    rememberMe: "Keep me signed in",
    signInBtn: "Sign In to Official Command →",
    signOut: "Sign Out",
    evalPersonas: "Or evaluate with a municipal role profile:"
  },
  am: {
    citizenPortal: "የተገልጋይ ማዕከል",
    officialDesk: "የኃላፊዎች መቆጣጠሪያ",
    activeServiceStatus: "የአዲስ አበባ ከተማ አስተዳደር ዲጂታል በር",
    heroTitle: "ቀልጣፋና ግልጽ የህዝብ አገልግሎት እና አውቶማቲክ ቅሬታ መፍቻ",
    heroDesc: "የወረፋና ቢሮክራሲ ችግርን በማስቀረት ቅሬታዎን በቀጥታ ለወረዳ ያቅርቡ፤ የጊዜ ሰሌዳውን (SLA) ይከታተሉ፤ እንዲሁም በዕሮብ እና አርብ የዲጂታል ችሎት ላይ በቀጥታ ይሳተፉ።",
    submitGrievance: "ቅሬታ ማቅረቢያ",
    trackTicket: "ቲኬትና QR መከታተያ",
    activeCasesCount: "በሂደት ላይ ያሉ",
    breachedCount: "እየተሸጋገሩ ያሉ",
    hearingDays: "የችሎት ቀናት (ዕሮብ/አርብ)",
    newComplaint: "አዲስ ቅሬታ ማስገቢያ",
    trackExisting: "የጉዳይ ሁኔታ መከታተያ",
    intakeFormTitle: "የተገልጋይ ቅሬታ ቅጽ",
    intakeFormSubtitle: "የጉዳይዎን ዝርዝር በመሙላት ህጋዊ መከታተያ ቲኬት እና የጊዜ ገደብ ወዲያውኑ ይቀበሉ።",
    fullName: "ሙሉ ስም *",
    phoneNumber: "ስልክ ቁጥር (ለኤስኤምኤስ መረጃ) *",
    nationalId: "ብሔራዊ መታወቂያ (ፋይዳ / ቀበሌ)",
    houseNumber: "የቤት ቁጥር / ቀበሌ",
    subCity: "ክፍለ ከተማ *",
    woreda: "የሚመለከተው ወረዳ *",
    serviceType: "የአገልግሎት ዘርፍ እና የተቀመጠው የሰዓት ገደብ *",
    priority: "የጉዳዩ አጣዳፊነት",
    grievanceTitle: "የጉዳዩ ዋና ርዕስ *",
    detailedDescription: "ዝርዝር አቤቱታ እና የቀድሞ ምላሾች *",
    evidenceUpload: "ማስረጃ ሰነዶች (ውል፣ ውድቅ ማስታወሻ፣ የድምፅ መልእክት)",
    submitGrievanceBtn: "ቅሬታውን መዝግብና ቲኬት አውጣ",
    trackBtn: "ጉዳዩን ፈልግ",
    demoTickets: "የሙከራ ቲኬቶች፦",
    triggerSla: "የ SLA አውቶማቲክ ማሸጋገሪያን አንቀሳቅስ",
    kpiTotal: "አጠቃላይ የቀረቡ ቅሬታዎች",
    kpiActive: "በሂደት ላይ ያሉ ጉዳዮች",
    kpiBreached: "ጊዜ ያለፈባቸው (Breached)",
    kpiEscalated: "ወደ ክ/ከተማ ወይም ከተማ የሸጋገሩ",
    kpiResolved: "መፍትሄ ያገኙ ጉዳዮች",
    grievanceQueue: "የጉዳዮች ዝርዝር ማዕከል",
    hearingDesk: "የዕሮብ እና አርብ ችሎቶች",
    bottleneckHeatmap: "የአስተዳደር ክፍተቶች መረጃ (Heatmap)",
    auditLedger: "የማይለወጥ የታሪክ መዝገብ (Audit)",
    hearingTitle: "የዲጂታል ችሎት ማዕከል (ዕሮብ እና አርብ)",
    hearingSubtitle: "በቀጥታ ቪዲዮ ውይይት ከኃላፊዎች ጋር በመገናኘት ውሳኔ የሚያገኙበት መድረክ።",
    selectHearingDay: "የችሎት ቀን ይምረጡ፦",
    voiceStudioTitle: "የተገልጋይ የድምጽ መልእክት መቅረጫ (Voice Memo)",
    voiceStudioDesc: "ቅሬታዎን በአማርኛ ወይም በእንግሊዝኛ በድምጽ ያስረዱ። ስርዓቱ በራስ-ሰር ወደ ጽሑፍ ይቀይረዋል።",
    startRecording: "ድምጽ ይቅረጹ",
    stopRecording: "ቅረጻውን አቁም",
    step1Title: "1. የተገልጋይ መረጃ",
    step2Title: "2. አስተዳደር እና ሰዓት",
    step3Title: "3. ዝርዝር ቅሬታ እና ማስረጃ",
    nextStep: "ቀጣይ ደረጃ →",
    prevStep: "← ወደ ኋላ",
    systemOnline: "የአዲስ አበባ ማዘጋጃ ቤት መስመር • ክፍት",
    staffManagement: "የሰራተኞች እና የስልጣን ማዕከል",
    loginTitle: "የአስተዳደር ሰራተኞች መግቢያ",
    loginSubtitle: "የአንድ ማዕከል ዲጂታል ችሎትና መቆጣጠሪያ",
    loginDesc: "ለተፈቀደላቸው ሰራተኞች ብቻ። የችሎት ቀጠሮዎችን ለማስተዳደር፣ የተሸጋገሩ አቤቱታዎችን ለመመርመር እና ህጋዊ ውሳኔዎችን ለማጽደቅ ይግቡ።",
    emailAddress: "ኦፊሴላዊ የኢሜይል አድራሻ *",
    password: "የይለፍ ቃል *",
    rememberMe: "በዚህ ኮምፒውተር ላይ እንዳለሁ ይቆይ",
    signInBtn: "ወደ መቆጣጠሪያው ግባ →",
    signOut: "ውጣ",
    evalPersonas: "ወይም በሙከራ የስራ መደብ ይግቡ፦"
  }
};

// ============================================================================
// THEME SWITCHING (DARK OBSIDIAN / CIVIC LIGHT)
// ============================================================================
function initTheme() {
  const saved = localStorage.getItem('civic_theme') || 'dark';
  document.documentElement.setAttribute('data-theme', saved);
  updateThemeButton(saved);
}

function toggleTheme() {
  const current = document.documentElement.getAttribute('data-theme') || 'dark';
  const next = current === 'dark' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', next);
  localStorage.setItem('civic_theme', next);
  updateThemeButton(next);
  showToast(`Switched theme to ${next === 'dark' ? 'Midnight Obsidian' : 'Civic Slate Light'}`, 'info');
}

function updateThemeButton(theme) {
  const icon = document.getElementById('themeToggleIcon');
  const label = document.getElementById('themeToggleLabel');
  if (icon) icon.textContent = theme === 'dark' ? '☀️' : '🌙';
  if (label) label.textContent = theme === 'dark' ? 'Light' : 'Dark';
}

// ============================================================================
// MULTI-STEP CITIZEN INTAKE WIZARD
// ============================================================================
let currentFormStep = 1;
const totalFormSteps = 3;

function initFormStepper() {
  updateFormStepUI();
}

function setFormStep(step) {
  if (step < 1 || step > totalFormSteps) return;
  if (step > currentFormStep) {
    if (!validateCurrentStep(currentFormStep)) return;
  }
  currentFormStep = step;
  updateFormStepUI();
}

function nextFormStep() {
  if (currentFormStep < totalFormSteps) {
    if (!validateCurrentStep(currentFormStep)) return;
    currentFormStep++;
    updateFormStepUI();
  }
}

function prevFormStep() {
  if (currentFormStep > 1) {
    currentFormStep--;
    updateFormStepUI();
  }
}

function validateCurrentStep(step) {
  if (step === 1) {
    const name = document.getElementById('citizenName').value.trim();
    const phone = document.getElementById('citizenPhone').value.trim();
    if (!name) {
      showToast('Please enter your full name', 'warning');
      document.getElementById('citizenName').focus();
      return false;
    }
    if (!phone) {
      showToast('Please enter your phone number for SLA alerts', 'warning');
      document.getElementById('citizenPhone').focus();
      return false;
    }
  } else if (step === 2) {
    const subCity = document.getElementById('subCitySelect').value;
    const woreda = document.getElementById('woredaSelect').value;
    const serviceType = document.getElementById('serviceTypeSelect').value;
    if (!subCity) {
      showToast('Please select your Sub-City Administration', 'warning');
      document.getElementById('subCitySelect').focus();
      return false;
    }
    if (!woreda) {
      showToast('Please select your Woreda Jurisdiction', 'warning');
      document.getElementById('woredaSelect').focus();
      return false;
    }
    if (!serviceType) {
      showToast('Please select a Public Service Classification', 'warning');
      document.getElementById('serviceTypeSelect').focus();
      return false;
    }
  }
  return true;
}

function updateFormStepUI() {
  for (let i = 1; i <= totalFormSteps; i++) {
    const ind = document.getElementById(`formStepIndicator${i}`);
    const sec = document.getElementById(`formStepSection${i}`);
    if (ind) {
      ind.classList.toggle('active', i === currentFormStep);
      ind.classList.toggle('completed', i < currentFormStep);
    }
    if (sec) {
      sec.classList.toggle('active', i === currentFormStep);
    }
  }

  const prevBtn = document.getElementById('formPrevBtn');
  const nextBtn = document.getElementById('formNextBtn');
  const submitBtn = document.getElementById('submitGrievanceBtn');

  if (prevBtn) prevBtn.style.display = currentFormStep > 1 ? 'inline-flex' : 'none';
  if (nextBtn) nextBtn.style.display = currentFormStep < totalFormSteps ? 'inline-flex' : 'none';
  if (submitBtn) submitBtn.style.display = currentFormStep === totalFormSteps ? 'inline-flex' : 'none';
}

// Initialize Application
document.addEventListener('DOMContentLoaded', async () => {
  initTheme();
  initFormStepper();
  await loadMunicipalHierarchy();
  await loadServiceTypes();
  await checkExistingSession();
  await checkGatewayHealth();
  initDatePickerDefaults();
});

// Set Bilingual Language
function setLanguage(lang) {
  currentLang = lang;
  document.getElementById('langEnBtn').classList.toggle('active', lang === 'en');
  document.getElementById('langAmBtn').classList.toggle('active', lang === 'am');

  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    if (i18n[lang] && i18n[lang][key]) {
      el.textContent = i18n[lang][key];
    }
  });

  if (lang === 'am') {
    document.body.classList.add('amharic-mode');
  } else {
    document.body.classList.remove('amharic-mode');
  }
}

// Switch between Citizen and Official Portals
function switchPortal(portal) {
  document.getElementById('modeCitizenBtn').classList.toggle('active', portal === 'citizen');
  document.getElementById('modeOfficialBtn').classList.toggle('active', portal === 'official');

  document.getElementById('citizenSection').classList.toggle('active', portal === 'citizen');
  document.getElementById('officialSection').classList.toggle('active', portal === 'official');

  if (portal === 'official') {
    if (currentToken && currentUser) {
      updateAuthUI(true);
      refreshOfficialDashboard();
    } else {
      updateAuthUI(false);
    }
  }
}

// Switch Citizen Tabs
function showCitizenTab(tab) {
  document.getElementById('citizenTabIntake').classList.toggle('active', tab === 'intake');
  document.getElementById('citizenTabTrack').classList.toggle('active', tab === 'track');

  document.getElementById('citizenIntakeTab').classList.toggle('active', tab === 'intake');
  document.getElementById('citizenTrackTab').classList.toggle('active', tab === 'track');
}

// Switch Official Tabs
function showOfficialTab(tab) {
  const tabs = ['cases', 'hearings', 'heatmap', 'audit', 'users'];
  tabs.forEach(t => {
    const btn = document.getElementById('offTab' + capitalize(t));
    const content = document.getElementById('off' + capitalize(t) + 'Tab');
    if (btn) btn.classList.toggle('active', t === tab);
    if (content) content.classList.toggle('active', t === tab);
  });
  if (tab === 'users') {
    loadOfficialUsers();
  }
}

function capitalize(s) {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

// Load Municipal Hierarchy (City -> Sub-Cities -> Woredas)
async function loadMunicipalHierarchy() {
  try {
    const res = await fetch(apiUrl('/api/v1/structures/tree'));
    if (!res.ok) return;
    municipalTree = await res.json();

    const subCitySelect = document.getElementById('subCitySelect');
    subCitySelect.innerHTML = '<option value="">-- Select Sub-City --</option>';

    if (municipalTree.length > 0 && municipalTree[0].children) {
      municipalTree[0].children.forEach(sc => {
        const opt = document.createElement('option');
        opt.value = sc.id;
        opt.textContent = sc.name;
        subCitySelect.appendChild(opt);
      });
    }
  } catch (err) {
    console.error("Failed to load hierarchy:", err);
  }
}

// On Sub-City Select -> Populate Child Woredas
function onSubCityChange() {
  const scID = parseInt(document.getElementById('subCitySelect').value);
  const woredaSelect = document.getElementById('woredaSelect');
  woredaSelect.innerHTML = '<option value="">-- Select Woreda --</option>';

  if (!scID || !municipalTree.length) return;

  const root = municipalTree[0];
  const subCity = root.children?.find(c => c.id === scID);
  if (subCity && subCity.children) {
    subCity.children.forEach(w => {
      const opt = document.createElement('option');
      opt.value = w.id;
      opt.textContent = w.name;
      woredaSelect.appendChild(opt);
    });
  }
}

// Load Service Types & Base SLAs
async function loadServiceTypes() {
  try {
    const res = await fetch(apiUrl('/api/v1/service-types'));
    if (!res.ok) return;
    serviceTypes = await res.json();

    const sel = document.getElementById('serviceTypeSelect');
    sel.innerHTML = '<option value="">-- Select Public Service Category --</option>';
    serviceTypes.forEach(st => {
      const opt = document.createElement('option');
      opt.value = st.id;
      opt.textContent = `${st.name} (${st.base_sla_hours}h SLA)`;
      opt.dataset.sla = st.base_sla_hours;
      sel.appendChild(opt);
    });
  } catch (err) {
    console.error("Failed to load service types:", err);
  }
}

function updateSlaBadge() {
  const sel = document.getElementById('serviceTypeSelect');
  const opt = sel.options[sel.selectedIndex];
  const badge = document.getElementById('slaTargetBadge');
  const hrsSpan = document.getElementById('slaTargetHours');

  if (opt && opt.dataset.sla) {
    badge.classList.remove('hidden');
    hrsSpan.textContent = opt.dataset.sla;
  } else {
    badge.classList.add('hidden');
  }
}

// Handle File Attachments
function handleFileSelect(e) {
  const files = e.target.files;
  if (!files || files.length === 0) return;

  for (let i = 0; i < files.length; i++) {
    const f = files[i];
    attachedFiles.push({
      file_name: f.name,
      file_url: `/uploads/${f.name}`,
      mime_type: f.type || 'application/octet-stream',
      extracted_ocr_text: `Simulated OCR text extracted from ${f.name}: Formal complaint and title deed reference confirmed.`
    });
  }
  renderFileTags();
}

function renderFileTags() {
  const list = document.getElementById('selectedFilesList');
  list.innerHTML = '';
  attachedFiles.forEach((f, idx) => {
    const tag = document.createElement('div');
    tag.className = 'file-tag';
    tag.innerHTML = `<span>📎 ${f.file_name}</span> <span style="cursor:pointer; color:#ef4444;" onclick="removeFile(${idx})">&times;</span>`;
    list.appendChild(tag);
  });
}

function removeFile(idx) {
  attachedFiles.splice(idx, 1);
  renderFileTags();
}

// Voice Memo Recording Simulator
let isRecordingVoice = false;
let voiceRecordSeconds = 0;
let voiceRecordInterval = null;

function toggleVoiceRecording() {
  const btn = document.getElementById('recordVoiceBtn');
  const icon = document.getElementById('recordVoiceIcon');
  const text = document.getElementById('recordVoiceText');
  const wave = document.getElementById('voiceWaveform');
  const timer = document.getElementById('recordingTimer');

  if (!isRecordingVoice) {
    isRecordingVoice = true;
    voiceRecordSeconds = 0;
    icon.textContent = '⏹️';
    text.textContent = (currentLang === 'am') ? 'ቅረጻውን አቁም' : 'Stop & Save Memo';
    btn.className = 'btn btn-warning btn-sm';
    wave.classList.remove('hidden');

    voiceRecordInterval = setInterval(() => {
      voiceRecordSeconds++;
      const m = String(Math.floor(voiceRecordSeconds / 60)).padStart(2, '0');
      const s = String(voiceRecordSeconds % 60).padStart(2, '0');
      timer.textContent = `${m}:${s}`;
    }, 1000);
  } else {
    isRecordingVoice = false;
    clearInterval(voiceRecordInterval);
    icon.textContent = '🔴';
    text.textContent = (currentLang === 'am') ? 'የድምጽ መልእክት ቅረጽ' : 'Record Voice Memo';
    btn.className = 'btn btn-secondary btn-sm';
    wave.classList.add('hidden');

    const audioFileName = `voice_memo_${Date.now()}.mp3`;
    attachedFiles.push({
      file_name: audioFileName,
      file_url: `/uploads/${audioFileName}`,
      mime_type: 'audio/mp3',
      extracted_ocr_text: `[Audio Voice Note Transcribed]: Citizen grievance regarding municipal delay and document dispute (${voiceRecordSeconds}s recording).`
    });
    renderFileTags();
    showToast((currentLang === 'am') ? 'የድምጽ መልእክትዎ በተሳካ ሁኔታ ተያይዟል!' : 'Voice memo recorded and attached to grievance!', 'success');
  }
}

function copyTicketNumber(tkt) {
  navigator.clipboard.writeText(tkt);
  showToast((currentLang === 'am') ? `ቲኬት ቁጥር ${tkt} ተገልብጧል!` : `Ticket number ${tkt} copied to clipboard!`, 'success');
}


// Submit Grievance (Citizen Intake)
async function handleGrievanceSubmit(e) {
  e.preventDefault();

  const submitBtn = document.getElementById('submitGrievanceBtn');
  submitBtn.disabled = true;
  submitBtn.innerHTML = '<span>⏳ Processing Intake...</span>';

  const payload = {
    citizen_full_name: document.getElementById('citizenName').value,
    citizen_phone: document.getElementById('citizenPhone').value,
    citizen_national_id: document.getElementById('citizenNationalId').value,
    citizen_house_number: document.getElementById('citizenHouseNo').value,
    woreda_id: parseInt(document.getElementById('woredaSelect').value),
    service_type_id: parseInt(document.getElementById('serviceTypeSelect').value),
    priority: document.getElementById('prioritySelect').value,
    title: document.getElementById('grievanceTitle').value,
    description: document.getElementById('grievanceDescription').value,
    attachments: attachedFiles
  };

  try {
    const res = await fetch(apiUrl('/api/v1/cases'), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Idempotency-Key': 'intake-' + Date.now()
      },
      body: JSON.stringify(payload)
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to submit grievance');
    }

    const created = await res.json();
    showToast(`Grievance ticket created: ${created.ticket_number}`, 'success');

    // Reset Form
    document.getElementById('grievanceForm').reset();
    setFormStep(1);
    attachedFiles = [];
    renderFileTags();
    document.getElementById('slaTargetBadge').classList.add('hidden');

    // Switch to Track Tab and search this ticket
    showCitizenTab('track');
    document.getElementById('trackInput').value = created.ticket_number;
    await renderTicketResult(created.ticket_number);

  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    submitBtn.disabled = false;
    submitBtn.innerHTML = '<span class="btn-icon">🚀</span> <span>Submit Grievance & Generate Ticket</span>';
  }
}

// Track Ticket & QR
async function searchTicket() {
  const query = document.getElementById('trackInput').value.trim();
  if (!query) {
    showToast('Please enter a ticket number or QR code', 'warning');
    return;
  }
  await renderTicketResult(query);
}

function quickSearch(tkt) {
  document.getElementById('trackInput').value = tkt;
  renderTicketResult(tkt);
}

async function renderTicketResult(query) {
  const container = document.getElementById('trackResultContainer');
  container.classList.remove('hidden');
  container.innerHTML = '<div style="text-align:center; padding:2rem;">⏳ Retrieving official ticket record...</div>';

  try {
    let res;
    if (query.startsWith('ETH-MUNI-')) {
      res = await fetch(apiUrl(`/api/v1/cases/verify/${encodeURIComponent(query)}`));
    } else {
      res = await fetch(apiUrl(`/api/v1/cases/ticket/${encodeURIComponent(query)}`));
    }

    if (!res.ok) {
      container.innerHTML = `
        <div class="track-detail-card" style="text-align:center;">
          <h3 style="color:#ef4444; margin-bottom:0.5rem;">❌ Ticket Not Found</h3>
          <p style="color:#94a3b8;">No registered grievance matches query: <strong>${query}</strong></p>
        </div>`;
      return;
    }

    const data = await res.json();
    renderDetailedTicketCard(data, container);

  } catch (err) {
    container.innerHTML = `<div class="track-detail-card" style="color:#ef4444;">Error: ${err.message}</div>`;
  }
}

function renderDetailedTicketCard(c, container) {
  const isBreached = c.is_breached;
  const status = c.status;

  // Compute status timeline active step
  const steps = [
    { label: "Submitted", code: "SUBMITTED" },
    { label: "Under Review", code: "UNDER_REVIEW" },
    { label: "Hearing Desk", code: "SCHEDULED_FOR_HEARING" },
    { label: "Resolved", code: "RESOLVED" }
  ];

  let stepHtml = '';
  steps.forEach((st, idx) => {
    let stateClass = '';
    if (status === 'RESOLVED') {
      stateClass = 'completed';
    } else if (status === 'ESCALATED_TO_SUBCITY' || status === 'ESCALATED_TO_CITY') {
      stateClass = (idx <= 1) ? 'escalated' : '';
    } else if (status === st.code) {
      stateClass = 'current';
    } else if (
      (status === 'UNDER_REVIEW' && idx === 0) ||
      (status === 'SCHEDULED_FOR_HEARING' && idx <= 1)
    ) {
      stateClass = 'completed';
    }

    stepHtml += `
      <div class="timeline-step ${stateClass}">
        <div class="step-circle">${idx + 1}</div>
        <div class="step-label">${st.label}</div>
      </div>
    `;
  });

  // Hearing slot markup
  let hearingBanner = '';
  if (c.hearing_slots && c.hearing_slots.length > 0) {
    const hs = c.hearing_slots[0];
    hearingBanner = `
      <div class="virtual-hearing-alert">
        <div class="hearing-alert-info">
          <h4>📅 Presiding Virtual Hearing Confirmed</h4>
          <p>Date: <strong>${hs.hearing_date}</strong> (10:00 - 10:30) with <strong>${hs.official_name}</strong></p>
          <p style="font-size:0.8rem; margin-top:0.25rem;">Meeting instructions: ${hs.hearing_notes || 'Please have your original documentation ready.'}</p>
        </div>
        <a href="${hs.meeting_link}" target="_blank" class="btn btn-primary btn-sm">
          <span>🎥 Enter Digital Hearing Room</span>
        </a>
      </div>
    `;
  }

  // Resolution markup
  let resolutionBox = '';
  if (c.status === 'RESOLVED') {
    resolutionBox = `
      <div class="resolution-certified-box">
        <h4>🛡️ Official Administrative Resolution & Ruling</h4>
        <p style="font-size:0.95rem; margin-bottom:0.5rem;">${c.resolution_summary || 'Resolved per municipal guidelines.'}</p>
        <div style="font-size:0.75rem; color:#94a3b8;">
          Presiding Officer: <strong>${c.resolved_by_name || 'Municipal Director'}</strong> • Timestamp: ${c.resolved_at ? new Date(c.resolved_at).toLocaleString() : 'Certified'}
        </div>
      </div>
    `;
  }

  // SLA clock display
  let slaClockClass = isBreached ? 'text-red' : 'text-green';
  let slaClockLabel = isBreached ? '⚠️ Mandated SLA Breached (Auto-Escalating)' : '⏱️ Active SLA Countdown';

  container.innerHTML = `
    <div class="track-detail-card">
      <div class="ticket-header-row">
        <div>
          <div class="ticket-num-badge">
            <span>🎫</span> ${c.ticket_number}
            <button type="button" class="copy-ticket-btn" onclick="copyTicketNumber('${c.ticket_number}')" title="Copy Ticket #">
              <span>📋 Copy</span>
            </button>
          </div>
          <div style="margin-top:0.3rem;">
            <span class="status-badge status-${c.status}">${c.status.replace(/_/g, ' ')}</span>
            <span style="font-size:0.8rem; color:#94a3b8; margin-left:0.5rem;">Escalation Tier: <strong>${c.current_structure_name} (${c.structure_level})</strong></span>
          </div>
        </div>

        <div class="sla-countdown-box">
          <div class="sla-clock-icon">${isBreached ? '🚨' : '⏱️'}</div>
          <div>
            <div class="sla-clock-text">${slaClockLabel}</div>
            <div class="sla-clock-time ${slaClockClass}" id="ticketCountdownDisplay">Calculating...</div>
          </div>
        </div>
      </div>

      <!-- Progression Timeline -->
      <div class="status-timeline">
        ${stepHtml}
      </div>

      ${hearingBanner}
      ${resolutionBox}

      <!-- Detailed Info & QR Verification -->
      <div class="track-info-grid">
        <div>
          <h3 style="font-size:1.15rem; margin-bottom:0.75rem;">${c.title}</h3>
          <p style="font-size:0.9rem; color:#cbd5e1; margin-bottom:1.25rem;">${c.description}</p>

          <div class="info-row">
            <div class="info-label">Applicant:</div>
            <div class="info-value">${c.citizen_name} (${c.citizen_phone})</div>
          </div>
          <div class="info-row">
            <div class="info-label">Classification:</div>
            <div class="info-value">${c.service_type_name}</div>
          </div>
          <div class="info-row">
            <div class="info-label">Jurisdiction Tier:</div>
            <div class="info-value">${c.current_structure_name} [Level: ${c.structure_level}]</div>
          </div>
          <div class="info-row">
            <div class="info-label">Assigned Officer:</div>
            <div class="info-value">${c.assigned_to_name || 'Unassigned (General Desk Queue)'}</div>
          </div>
          <div class="info-row">
            <div class="info-label">Lodged At:</div>
            <div class="info-value">${new Date(c.created_at).toLocaleString()}</div>
          </div>
        </div>

        <div class="qr-box">
          <div id="qrCodeCanvas"></div>
          <div class="qr-hash-text">${c.qr_verification_code}</div>
          <div style="font-size:0.7rem; color:#666; margin-top:0.25rem;">Scan to verify authenticity</div>
        </div>
      </div>

      <!-- Audit History -->
      <div style="margin-top:2rem; border-top:1px solid rgba(255,255,255,0.08); padding-top:1.5rem;">
        <h4 style="font-size:0.95rem; margin-bottom:1rem; color:#94a3b8;">📜 Cryptographic Audit Ledger Trail (${c.audit_logs?.length || 0} events)</h4>
        <div style="display:flex; flex-direction:column; gap:0.6rem;">
          ${(c.audit_logs || []).map(l => `
            <div style="background:rgba(10,15,29,0.5); padding:0.65rem 0.85rem; border-radius:8px; font-size:0.8rem; border:1px solid rgba(255,255,255,0.05);">
              <div style="display:flex; justify-content:space-between; font-weight:600; color:#3b82f6;">
                <span>${l.action}</span>
                <span style="color:#64748b; font-family:var(--font-mono);">${new Date(l.created_at).toLocaleTimeString()}</span>
              </div>
              <div style="color:#94a3b8; margin-top:0.15rem;">${l.notes || ''} (Actor: ${l.performer_name})</div>
            </div>
          `).join('')}
        </div>
      </div>
    </div>
  `;

  // Render QR Code
  setTimeout(() => {
    const qrDiv = document.getElementById('qrCodeCanvas');
    if (qrDiv) {
      qrDiv.innerHTML = '';
      new QRCode(qrDiv, {
        text: window.location.origin + '/api/v1/cases/verify/' + c.qr_verification_code,
        width: 130,
        height: 130,
        colorDark: "#000000",
        colorLight: "#ffffff"
      });
    }
  }, 50);

  // Start live timer
  startTicketCountdown(c.sla_deadline, c.status);
}

function startTicketCountdown(deadlineIso, status) {
  if (activeCountdownInterval) clearInterval(activeCountdownInterval);

  function update() {
    const el = document.getElementById('ticketCountdownDisplay');
    if (!el) return;

    if (status === 'RESOLVED') {
      el.textContent = 'RESOLVED';
      el.className = 'sla-clock-time text-green';
      return;
    }

    const now = new Date().getTime();
    const target = new Date(deadlineIso).getTime();
    const diff = target - now;

    if (diff <= 0) {
      const overHrs = Math.floor(Math.abs(diff) / 3600000);
      const overMins = Math.floor((Math.abs(diff) % 3600000) / 60000);
      el.textContent = `BREACHED by +${overHrs}h ${overMins}m`;
      el.className = 'sla-clock-time text-red';
    } else {
      const hrs = Math.floor(diff / 3600000);
      const mins = Math.floor((diff % 3600000) / 60000);
      const secs = Math.floor((diff % 60000) / 1000);
      el.textContent = `${hrs}h ${mins}m ${secs}s`;
      el.className = (hrs < 6) ? 'sla-clock-time text-red' : 'sla-clock-time text-green';
    }
  }

  update();
  activeCountdownInterval = setInterval(update, 1000);
}

// ============================================================================
// OFFICIAL COMMAND CENTER OPERATIONS
// ============================================================================
// OFFICIAL COMMAND CENTER OPERATIONS & SESSION MANAGEMENT
// ============================================================================

// Check and validate active session from localStorage
async function checkExistingSession() {
  const savedToken = localStorage.getItem(SESSION_TOKEN_KEY);
  if (!savedToken) {
    updateAuthUI(false);
    return false;
  }

  try {
    const res = await fetch(apiUrl('/api/v1/auth/me'), {
      headers: { 'Authorization': 'Bearer ' + savedToken }
    });

    if (res.ok) {
      const user = await res.json();
      currentToken = savedToken;
      currentUser = user;
      updateAuthUI(true);
      await refreshOfficialDashboard();
      return true;
    } else {
      // Token expired or invalid
      logoutOfficial(false);
      return false;
    }
  } catch (err) {
    console.warn("Session validation error:", err);
    updateAuthUI(false);
    return false;
  }
}

// Update UI elements depending on auth state
function updateAuthUI(isAuthenticated) {
  const loginGate = document.getElementById('officialLoginGate');
  const dashView = document.getElementById('officialDashboardView');
  const navSessionBadge = document.getElementById('navSessionBadge');
  const navSessionName = document.getElementById('navSessionName');
  const navSessionAvatar = document.getElementById('navSessionAvatar');

  if (isAuthenticated && currentUser) {
    if (loginGate) loginGate.classList.add('hidden');
    if (dashView) dashView.classList.remove('hidden');

    // Nav session pill
    if (navSessionBadge) navSessionBadge.classList.remove('hidden');
    if (navSessionName) navSessionName.textContent = `${currentUser.full_name} (${currentUser.role})`;

    // Header profile card
    const nameEl = document.getElementById('officerNameDisplay');
    const roleEl = document.getElementById('officerRoleBadge');
    const structEl = document.getElementById('officerStructBadge');
    const avatarEl = document.getElementById('officerAvatar');

    if (nameEl) nameEl.textContent = currentUser.full_name;
    if (roleEl) roleEl.textContent = currentUser.role;
    if (structEl) structEl.textContent = currentUser.structure_name || 'City Administration';

    let avatar = '⚖️';
    if (currentUser.role === 'SUPER_ADMIN') avatar = '🛡️';
    else if (currentUser.role === 'CITY_DIRECTOR') avatar = '⚖️';
    else if (currentUser.role === 'SUBCITY_MANAGER') avatar = '🏢';
    else if (currentUser.role === 'WOREDA_OFFICER') avatar = '🏛️';

    if (avatarEl) avatarEl.textContent = avatar;
    if (navSessionAvatar) navSessionAvatar.textContent = avatar;

    // Role switcher dropdown synchronization
    const roleSelect = document.getElementById('roleSwitchSelect');
    if (roleSelect && currentUser.email) {
      for (let i = 0; i < roleSelect.options.length; i++) {
        if (roleSelect.options[i].value === currentUser.email) {
          roleSelect.selectedIndex = i;
          break;
        }
      }
    }
  } else {
    if (loginGate) loginGate.classList.remove('hidden');
    if (dashView) dashView.classList.add('hidden');
    if (navSessionBadge) navSessionBadge.classList.add('hidden');
  }
}

// Handle official form submission login
async function handleOfficialLogin(e) {
  if (e) e.preventDefault();

  const emailInput = document.getElementById('loginEmail');
  const pwdInput = document.getElementById('loginPassword');
  const submitBtn = document.getElementById('loginSubmitBtn');
  const errorAlert = document.getElementById('loginErrorAlert');

  if (errorAlert) errorAlert.style.display = 'none';

  const email = emailInput ? emailInput.value.trim() : '';
  const password = pwdInput ? pwdInput.value : '';

  if (!email || !password) {
    showLoginError("Please enter both official email and password.");
    return;
  }

  if (submitBtn) {
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span>⏳ Authenticating credentials...</span>';
  }

  try {
    const res = await fetch(apiUrl('/api/v1/auth/login'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Authentication failed. Please check your credentials.');
    }

    // Login successful
    currentToken = data.token;
    currentUser = data.user;

    localStorage.setItem(SESSION_TOKEN_KEY, currentToken);
    localStorage.setItem(SESSION_USER_KEY, JSON.stringify(currentUser));

    updateAuthUI(true);
    showToast(`Welcome back, ${currentUser.full_name}! (${currentUser.role})`, 'success');

    if (pwdInput) pwdInput.value = '';

    await refreshOfficialDashboard();

  } catch (err) {
    showLoginError(err.message || 'Authentication failed. Check your password or role assignment.');
  } finally {
    if (submitBtn) {
      submitBtn.disabled = false;
      submitBtn.innerHTML = `<span class="btn-icon">🔐</span><span data-i18n="signInBtn">${i18n[currentLang]?.signInBtn || 'Sign In to Official Command →'}</span>`;
    }
  }
}

// Quick 1-click evaluation persona login
async function quickLogin(email) {
  const emailInput = document.getElementById('loginEmail');
  const pwdInput = document.getElementById('loginPassword');
  if (emailInput) emailInput.value = email;
  if (pwdInput) pwdInput.value = 'Password123!';

  showToast(`Authenticating evaluation persona: ${email}...`, 'info');
  await handleOfficialLogin(null);
}

// Logout official user and clear session
function logoutOfficial(showNotification = true) {
  localStorage.removeItem(SESSION_TOKEN_KEY);
  localStorage.removeItem(SESSION_USER_KEY);
  currentToken = '';
  currentUser = null;

  updateAuthUI(false);

  if (showNotification) {
    showToast('Official session terminated. Workstation locked.', 'info');
  }
}

// Toggle password text/password visibility
function togglePasswordVisibility() {
  const pwdInput = document.getElementById('loginPassword');
  const icon = document.getElementById('pwdToggleIcon');
  const label = document.getElementById('pwdToggleLabel');
  if (!pwdInput) return;

  if (pwdInput.type === 'password') {
    pwdInput.type = 'text';
    if (icon) icon.textContent = '🙈';
    if (label) label.textContent = 'Hide';
  } else {
    pwdInput.type = 'password';
    if (icon) icon.textContent = '👁️';
    if (label) label.textContent = 'Show';
  }
}

// Display login error banner
function showLoginError(msg) {
  const errorAlert = document.getElementById('loginErrorAlert');
  const errorMsg = document.getElementById('loginErrorMsg');
  if (errorAlert && errorMsg) {
    errorMsg.textContent = msg;
    errorAlert.style.display = 'flex';
  } else {
    showToast(msg, 'error');
  }
}

// Handle switching role account while authenticated
async function handleUserRoleSwitch(e) {
  const email = e.target.value;
  try {
    const res = await fetch(apiUrl('/api/v1/auth/login'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: email, password: 'Password123!' })
    });
    if (!res.ok) {
      showToast('Failed to switch persona account', 'error');
      return;
    }

    const data = await res.json();
    currentToken = data.token;
    currentUser = data.user;

    localStorage.setItem(SESSION_TOKEN_KEY, currentToken);
    localStorage.setItem(SESSION_USER_KEY, JSON.stringify(currentUser));

    updateAuthUI(true);
    showToast(`Switched account: ${currentUser.full_name} (${currentUser.role})`, 'success');
    await refreshOfficialDashboard();

  } catch (err) {
    showToast("Role switch failed: " + err.message, 'error');
  }
}

async function refreshOfficialDashboard() {
  await loadExecutiveKPIs();
  await loadOfficialCases();
  await loadHearingsForDate();
  await loadOfficialUsers();
}

async function loadExecutiveKPIs() {
  try {
    const res = await fetch(apiUrl('/api/v1/analytics/dashboard'), {
      headers: { 'Authorization': 'Bearer ' + currentToken }
    });
    if (!res.ok) return;

    const data = await res.json();
    document.getElementById('kpiTotal').textContent = data.total_cases;
    document.getElementById('kpiActive').textContent = data.active_cases;
    document.getElementById('kpiBreached').textContent = data.breached_cases;
    document.getElementById('kpiEscalated').textContent = data.escalated_to_subcity + data.escalated_to_city;
    document.getElementById('kpiResolved').textContent = data.resolved_cases;

    // Update Hero stats
    document.getElementById('heroActiveCases').textContent = data.active_cases;
    document.getElementById('heroBreachedCases').textContent = data.breached_cases;

    // Render Heatmap Table
    renderHeatmapTable(data.structure_breakdown || []);

    // Render Recent Audit Logs
    renderAuditLedger(data.recent_audit_logs || []);

  } catch (err) {
    console.error("Failed to load KPIs:", err);
  }
}

function renderHeatmapTable(breakdown) {
  const tbody = document.getElementById('heatmapTableBody');
  tbody.innerHTML = '';

  if (breakdown.length === 0) {
    tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;">No tier records found.</td></tr>';
    return;
  }

  breakdown.forEach(item => {
    const rate = item.case_count > 0 ? ((item.breached_count / item.case_count) * 100).toFixed(1) : 0;
    const efficiency = 100 - parseFloat(rate);
    const effClass = efficiency > 75 ? 'text-green' : (efficiency > 50 ? 'text-yellow' : 'text-red');

    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td><strong>${item.name}</strong></td>
      <td><span class="officer-struct-badge">${item.level}</span></td>
      <td>${item.case_count}</td>
      <td class="${item.breached_count > 0 ? 'text-red font-bold' : ''}">${item.breached_count}</td>
      <td>${rate}%</td>
      <td class="${effClass} font-bold">${efficiency.toFixed(1)}%</td>
    `;
    tbody.appendChild(tr);
  });
}

function renderAuditLedger(logs) {
  const container = document.getElementById('auditLogList');
  container.innerHTML = '';

  if (logs.length === 0) {
    container.innerHTML = '<div style="text-align:center; color:#94a3b8;">No audit records found.</div>';
    return;
  }

  logs.forEach(l => {
    const item = document.createElement('div');
    item.className = 'audit-item';
    item.innerHTML = `
      <div class="audit-item-icon">🛡️</div>
      <div style="flex:1;">
        <div class="audit-action-title">${l.action}</div>
        <div class="audit-meta">Performer: <strong>${l.performer_name}</strong> • ${new Date(l.created_at).toLocaleString()}</div>
        <div class="audit-notes">${l.notes || 'System action executed.'}</div>
      </div>
    `;
    container.appendChild(item);
  });
}

// Load Official Cases Grid
async function loadOfficialCases() {
  const status = document.getElementById('filterStatusSelect').value;
  const breached = document.getElementById('filterBreachSelect').value;

  let url = apiUrl('/api/v1/cases?limit=100');
  if (status) url += `&status=${status}`;
  if (breached) url += `&only_breached=${breached}`;

  try {
    const res = await fetch(url, {
      headers: { 'Authorization': 'Bearer ' + currentToken }
    });
    if (!res.ok) return;

    const data = await res.json();
    allCases = data.data || [];
    renderCasesTable(allCases);

  } catch (err) {
    console.error("Failed to load cases:", err);
  }
}

function renderCasesTable(cases) {
  const tbody = document.getElementById('casesTableBody');
  tbody.innerHTML = '';

  if (cases.length === 0) {
    tbody.innerHTML = '<tr><td colspan="7" style="text-align:center; padding:2rem; color:#94a3b8;">No grievance tickets matching criteria.</td></tr>';
    return;
  }

  cases.forEach(c => {
    const isBreached = c.is_breached;
    const remainingHrs = c.remaining_hours ? c.remaining_hours.toFixed(1) : 0;

    let slaTag = '';
    if (c.status === 'RESOLVED') {
      slaTag = `<span class="sla-tag normal">Completed</span>`;
    } else if (isBreached) {
      slaTag = `<span class="sla-tag breached">⚠️ Overdue (${Math.abs(remainingHrs)}h)</span>`;
    } else if (remainingHrs < 6) {
      slaTag = `<span class="sla-tag warning">${remainingHrs}h remaining</span>`;
    } else {
      slaTag = `<span class="sla-tag normal">${remainingHrs}h remaining</span>`;
    }

    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>
        <span class="ticket-cell" onclick="inspectCase('${c.ticket_number}')">${c.ticket_number}</span>
      </td>
      <td>
        <div><strong>${c.citizen_name}</strong></div>
        <div style="font-size:0.75rem; color:#94a3b8;">${c.citizen_phone}</div>
      </td>
      <td>${c.service_type_name}</td>
      <td>
        <div>${c.current_structure_name}</div>
        <div style="font-size:0.75rem; color:#3b82f6;">${c.structure_level}</div>
      </td>
      <td>${slaTag}</td>
      <td>
        <span class="status-badge status-${c.status}">${c.status.replace(/_/g, ' ')}</span>
      </td>
      <td>
        <div style="display:flex; gap:0.4rem;">
          <button class="btn btn-secondary btn-sm" title="Schedule Wed/Fri Hearing" onclick="openHearingModal('${c.id}', '${c.ticket_number}')">
            📅
          </button>
          <button class="btn btn-primary btn-sm" title="Resolve Case" onclick="openResolveModal('${c.id}', '${c.ticket_number}')">
            ✍️
          </button>
          <button class="btn btn-warning btn-sm" title="Manual Escalate" onclick="triggerManualEscalate('${c.id}')">
            ⚡
          </button>
        </div>
      </td>
    `;
    tbody.appendChild(tr);
  });
}

function filterCasesTable() {
  const q = document.getElementById('officialSearchInput').value.toLowerCase();
  const filtered = allCases.filter(c => 
    c.ticket_number.toLowerCase().includes(q) ||
    c.citizen_name.toLowerCase().includes(q) ||
    c.title.toLowerCase().includes(q)
  );
  renderCasesTable(filtered);
}

function inspectCase(ticket) {
  switchPortal('citizen');
  showCitizenTab('track');
  document.getElementById('trackInput').value = ticket;
  renderTicketResult(ticket);
}

// Trigger SLA Auto-Escalation Engine Sweep
async function triggerSLASweepNow() {
  const btn = document.getElementById('triggerSlaBtn');
  btn.disabled = true;
  btn.innerHTML = '<span>⚡ Running Row-Locked Sweep...</span>';

  try {
    const res = await fetch(apiUrl('/api/v1/sla/trigger-sweep'), {
      method: 'POST',
      headers: { 'Authorization': 'Bearer ' + currentToken }
    });

    if (!res.ok) throw new Error('SLA Sweep failed');

    const result = await res.json();
    if (result.escalated_count > 0) {
      showToast(`⚡ SLA Engine: ${result.escalated_count} breached tickets automatically escalated up municipal tiers!`, 'warning');
    } else {
      showToast('SLA Engine: Checked all tickets. No new SLA breaches detected.', 'success');
    }

    await refreshOfficialDashboard();

  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    btn.disabled = false;
    btn.innerHTML = '<span class="btn-icon">⚡</span> <span>Run SLA Escalation Engine</span>';
  }
}

// Manual Escalation
async function triggerManualEscalate(caseID) {
  const reason = prompt("Enter justification for administrative escalation to higher tier:");
  if (!reason) return;

  try {
    const res = await fetch(apiUrl(`/api/v1/cases/${caseID}/escalate`), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + currentToken
      },
      body: JSON.stringify({ reason })
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to escalate');
    }

    showToast('Case escalated successfully to higher municipal tier', 'success');
    await refreshOfficialDashboard();

  } catch (err) {
    showToast(err.message, 'error');
  }
}

// Modals: Hearings & Resolve
function openHearingModal(caseID, ticket) {
  document.getElementById('modalCaseId').value = caseID;
  document.getElementById('modalTicketNum').value = ticket;

  // Set default to next Wednesday
  const d = getNextWednesdayDate();
  document.getElementById('modalHearingDate').value = d;
  document.getElementById('modalMeetingLink').value = `https://meet.jit.si/eth-muni-hearing-${ticket}`;

  document.getElementById('scheduleHearingModal').classList.remove('hidden');
}

function openResolveModal(caseID, ticket) {
  document.getElementById('modalResolveCaseId').value = caseID;
  document.getElementById('modalResolveTicket').value = ticket;
  document.getElementById('resolveCaseModal').classList.remove('hidden');
}

async function handleScheduleHearingSubmit(e) {
  e.preventDefault();
  const caseID = document.getElementById('modalCaseId').value;
  const hearingDate = document.getElementById('modalHearingDate').value;
  const startTime = document.getElementById('modalStartTime').value;
  const endTime = document.getElementById('modalEndTime').value;
  const meetingLink = document.getElementById('modalMeetingLink').value;
  const notes = document.getElementById('modalHearingNotes').value;

  try {
    const res = await fetch(apiUrl(`/api/v1/cases/${caseID}/hearings`), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + currentToken
      },
      body: JSON.stringify({
        hearing_date: hearingDate,
        start_time: startTime,
        end_time: endTime,
        meeting_link: meetingLink,
        notes: notes
      })
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to schedule hearing');
    }

    showToast('Hearing appointment scheduled & citizen notified', 'success');
    closeModal('scheduleHearingModal');
    await refreshOfficialDashboard();

  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function handleResolveCaseSubmit(e) {
  e.preventDefault();
  const caseID = document.getElementById('modalResolveCaseId').value;
  const summary = document.getElementById('modalResolutionSummary').value;
  const notes = document.getElementById('modalDecisionNotes').value;

  try {
    const res = await fetch(apiUrl(`/api/v1/cases/${caseID}/resolve`), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + currentToken
      },
      body: JSON.stringify({
        resolution_summary: summary,
        decision_notes: notes,
        send_notification: true
      })
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to resolve case');
    }

    showToast('Case resolved and decision certified on ledger', 'success');
    closeModal('resolveCaseModal');
    await refreshOfficialDashboard();

  } catch (err) {
    showToast(err.message, 'error');
  }
}

// Load Hearings Desk for Date
async function loadHearingsForDate() {
  const picker = document.getElementById('hearingDatePicker');
  const dateStr = picker.value || getNextWednesdayDate();
  picker.value = dateStr;

  document.getElementById('hearingDateHeading').textContent = `Presiding Sessions for ${dateStr}`;

  try {
    const res = await fetch(apiUrl(`/api/v1/hearings/date/${dateStr}`), {
      headers: { 'Authorization': 'Bearer ' + currentToken }
    });
    if (!res.ok) return;

    const slots = await res.json();
    const container = document.getElementById('hearingSlotsList');
    document.getElementById('hearingSlotCount').textContent = `${slots.length} Sessions`;
    container.innerHTML = '';

    if (slots.length === 0) {
      container.innerHTML = '<div style="color:#94a3b8; padding:1.5rem; text-align:center;">No hearings scheduled for this date. (Wednesday & Friday desk open for bookings)</div>';
      return;
    }

    slots.forEach(s => {
      const card = document.createElement('div');
      card.className = 'hearing-slot-card';
      card.innerHTML = `
        <div>
          <div class="slot-time-badge">🕒 ${s.start_time} - ${s.end_time}</div>
          <div style="font-weight:600; font-size:1.05rem;">Case: ${s.case ? s.case.ticket_number : 'Hearing Slot'}</div>
          <div style="font-size:0.85rem; color:#94a3b8;">${s.hearing_notes || 'Citizen evidence presentation'}</div>
        </div>
        <div style="display:flex; gap:0.5rem;">
          <a href="${s.meeting_link}" target="_blank" class="btn btn-primary btn-sm">
            <span>🎥 Launch Room</span>
          </a>
        </div>
      `;
      container.appendChild(card);
    });

  } catch (err) {
    console.error("Failed to load hearings:", err);
  }
}

function initDatePickerDefaults() {
  const d = getNextWednesdayDate();
  const picker = document.getElementById('hearingDatePicker');
  if (picker) picker.value = d;
}

function getNextWednesdayDate() {
  const today = new Date();
  const day = today.getDay();
  let daysUntilWed = (3 - day + 7) % 7;
  if (daysUntilWed === 0) daysUntilWed = 7;
  const nextWed = new Date(today.getTime() + daysUntilWed * 86400000);
  return nextWed.toISOString().split('T')[0];
}

// Toast Notifications
function showToast(msg, type = 'info') {
  const container = document.getElementById('toastContainer');
  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.innerHTML = msg;
  container.appendChild(toast);

  setTimeout(() => {
    toast.remove();
  }, 4000);
}

// ============================================================================
// OFFICIAL STAFF & ADMINISTRATOR MANAGEMENT (ADD & REMOVE)
// ============================================================================
let allOfficialUsers = [];

async function loadOfficialUsers() {
  try {
    const res = await fetch(apiUrl('/api/v1/users'), {
      headers: { 'Authorization': 'Bearer ' + currentToken }
    });
    if (!res.ok) return;

    allOfficialUsers = await res.json();
    renderOfficialUsersTable(allOfficialUsers);
    updateRoleSwitcherOptions(allOfficialUsers);
  } catch (err) {
    console.error("Failed to load official users:", err);
  }
}

function renderOfficialUsersTable(users) {
  const tbody = document.getElementById('officialUsersTableBody');
  if (!tbody) return;
  tbody.innerHTML = '';

  if (!users || users.length === 0) {
    tbody.innerHTML = '<tr><td colspan="6" style="text-align:center; padding:2rem; color:var(--text-tertiary);">No official staff records found.</td></tr>';
    return;
  }

  users.forEach(u => {
    const tr = document.createElement('tr');
    const isCurrent = currentUser && currentUser.email === u.email;

    tr.innerHTML = `
      <td>
        <div style="display:flex; align-items:center; gap:0.75rem;">
          <div style="width:38px; height:38px; border-radius:50%; background:var(--bg-surface-elevated); border:1px solid var(--border-glass-bright); display:flex; align-items:center; justify-content:center; font-size:1.15rem;">
            ${getRoleAvatar(u.role)}
          </div>
          <div>
            <div style="font-weight:700;">${u.full_name} ${isCurrent ? '<span style="font-size:0.72rem; color:var(--cyan-400);">(You)</span>' : ''}</div>
            <div style="font-size:0.75rem; color:var(--text-tertiary); font-family:var(--font-mono);">ID: ${u.id.substring(0, 8)}...</div>
          </div>
        </div>
      </td>
      <td>
        <div><strong>${u.email}</strong></div>
        <div style="font-size:0.75rem; color:var(--text-tertiary);">${u.phone_number}</div>
      </td>
      <td>
        <span class="officer-role-badge">${u.role}</span>
      </td>
      <td>
        <div>${u.structure_name || 'Addis Ababa City Admin'}</div>
        <div style="font-size:0.75rem; color:var(--text-tertiary);">${u.admin_level || 'CITY'} Tier</div>
      </td>
      <td>
        <span style="display:inline-flex; align-items:center; gap:0.4rem; font-size:0.78rem; font-weight:700; color:var(--emerald-400); background:rgba(16,185,129,0.1); padding:0.25rem 0.65rem; border-radius:9999px; border:1px solid rgba(16,185,129,0.3);">
          <span style="width:6px; height:6px; border-radius:50%; background:var(--emerald-400);"></span> Active
        </span>
      </td>
      <td>
        ${isCurrent ? '<span style="font-size:0.75rem; color:var(--text-tertiary);">Active Session</span>' : `
          <button class="btn btn-secondary btn-sm" style="color:var(--rose-400); border-color:rgba(244,63,94,0.35);" onclick="deleteOfficialUser('${u.id}', '${u.full_name}')" title="Delete User">
            <span>🗑️ Remove</span>
          </button>
        `}
      </td>
    `;
    tbody.appendChild(tr);
  });
}

function getRoleAvatar(role) {
  switch (role) {
    case 'SUPER_ADMIN': return '👑';
    case 'CITY_DIRECTOR': return '🏛️';
    case 'SUBCITY_MANAGER': return '🏢';
    case 'WOREDA_OFFICER': return '⚖️';
    case 'SERVICE_DESK_AGENT': return '🎧';
    default: return '👤';
  }
}

function filterOfficialUsersTable() {
  const q = (document.getElementById('officialUserSearchInput')?.value || '').toLowerCase();
  const filtered = allOfficialUsers.filter(u =>
    u.full_name.toLowerCase().includes(q) ||
    u.email.toLowerCase().includes(q) ||
    u.role.toLowerCase().includes(q) ||
    (u.structure_name && u.structure_name.toLowerCase().includes(q))
  );
  renderOfficialUsersTable(filtered);
}

function updateRoleSwitcherOptions(users) {
  const select = document.getElementById('roleSwitchSelect');
  if (!select) return;
  const currentVal = select.value;
  select.innerHTML = '';

  users.forEach(u => {
    const opt = document.createElement('option');
    opt.value = u.email;
    opt.textContent = `${u.full_name} (${u.role.replace(/_/g, ' ')})`;
    if (u.email === currentVal) opt.selected = true;
    select.appendChild(opt);
  });
}

async function openCreateUserModal() {
  if (!municipalTree || municipalTree.length === 0) {
    await loadMunicipalHierarchy();
  }

  const sel = document.getElementById('newUserStructure');
  if (sel) {
    sel.innerHTML = '<option value="">-- Addis Ababa City Administration (Global) --</option>';

    if (municipalTree && municipalTree.length > 0 && municipalTree[0].children) {
      municipalTree[0].children.forEach(sc => {
        const scOpt = document.createElement('option');
        scOpt.value = sc.id;
        scOpt.textContent = `📍 [Sub-City] ${sc.name}`;
        sel.appendChild(scOpt);

        if (sc.children) {
          sc.children.forEach(w => {
            const wOpt = document.createElement('option');
            wOpt.value = w.id;
            wOpt.textContent = `  ↳ [Woreda] ${w.name}`;
            sel.appendChild(wOpt);
          });
        }
      });
    }
  }

  // Clear inputs for clean form
  const nameInp = document.getElementById('newUserName');
  const emailInp = document.getElementById('newUserEmail');
  const phoneInp = document.getElementById('newUserPhone');
  const pwdInp = document.getElementById('newUserPassword');
  const roleInp = document.getElementById('newUserRole');
  if (nameInp) nameInp.value = '';
  if (emailInp) emailInp.value = '';
  if (phoneInp) phoneInp.value = '';
  if (pwdInp) pwdInp.value = 'Password123!';
  if (roleInp) roleInp.value = 'WOREDA_OFFICER';

  openModal('createUserModal');
}

async function handleCreateUserSubmit(e) {
  e.preventDefault();
  const fullName = document.getElementById('newUserName').value.trim();
  const email = document.getElementById('newUserEmail').value.trim();
  const phone = document.getElementById('newUserPhone').value.trim();
  const password = document.getElementById('newUserPassword').value;
  const role = document.getElementById('newUserRole').value;
  const structVal = document.getElementById('newUserStructure').value;
  const structId = structVal ? parseInt(structVal) : null;

  if (!fullName || !email || !password || !role) {
    showToast('Please fill in all required official account details.', 'error');
    return;
  }

  const submitBtn = e.target.querySelector('button[type="submit"]');
  const origBtnText = submitBtn ? submitBtn.innerHTML : '';
  if (submitBtn) {
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span>⏳ Provisioning...</span>';
  }

  try {
    const res = await fetch(apiUrl('/api/v1/users'), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + currentToken
      },
      body: JSON.stringify({
        full_name: fullName,
        email: email,
        phone_number: phone,
        password: password,
        role: role,
        structure_id: structId
      })
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to create official user account');
    }

    const created = await res.json();
    showToast(`Official account created: ${created.full_name} (${created.role})`, 'success');
    closeModal('createUserModal');
    await loadOfficialUsers();

  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    if (submitBtn) {
      submitBtn.disabled = false;
      submitBtn.innerHTML = origBtnText;
    }
  }
}

async function deleteOfficialUser(userId, userName) {
  if (!confirm(`Are you sure you want to remove official staff account: ${userName}?`)) {
    return;
  }

  try {
    const res = await fetch(apiUrl(`/api/v1/users/${userId}`), {
      method: 'DELETE',
      headers: {
        'Authorization': 'Bearer ' + currentToken
      }
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to delete official user');
    }

    showToast(`Removed official account: ${userName}`, 'info');
    await loadOfficialUsers();

  } catch (err) {
    showToast(err.message, 'error');
  }
}
