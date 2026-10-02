// ==============================================================================
// SMART ONE-STOP MUNICIPAL CIVIC PLATFORM - ETHIOPIAN WORKING-DAY & CALENDAR ENGINE
// Pure Vanilla JavaScript Client-Side Engine (Bilingual EN / አማርኛ)
// ==============================================================================

(function(window) {
  'use strict';

  // Ethiopian Julian Day Number Epoch: Meskerem 1, 1 EE = JDN 1,724,221
  const ETHIOPIC_JDN_EPOCH = 1724221;

  // Month Names (1 to 13)
  const ETHIOPIAN_MONTHS_EN = [
    "", "Meskerem", "Tikimt", "Hidar", "Tahsas", "Tir", "Yekatit",
    "Megabit", "Miazia", "Ginbot", "Sene", "Hamle", "Nehase", "Pagume"
  ];

  const ETHIOPIAN_MONTHS_AM = [
    "", "መስከረም", "ጥቅምት", "ኅዳር", "ታኅሣሥ", "ጥር", "የካቲት",
    "መጋቢት", "ሚያዝያ", "ግንቦት", "ሰኔ", "ሐምሌ", "ነሐሴ", "ጳጉሜ"
  ];

  const GREGORIAN_MONTHS_EN = [
    "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"
  ];

  const GREGORIAN_MONTHS_AM = [
    "ጃንዩወሪ", "ፌብሩወሪ", "ማርች", "ኤፕሪል", "ሜይ", "ጁን", "ጁላይ", "ኦገስት", "ሴፕቴምበር", "ኦክቶበር", "ኖቬምበር", "ዲሴምበር"
  ];

  const WEEKDAYS_EN = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
  const WEEKDAYS_AM = ["እሑድ", "ሰኞ", "ማክሰኞ", "ረቡዕ", "ሐሙስ", "ዓርብ", "ቅዳሜ"];
  const WEEKDAYS_SHORT_EN = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
  const WEEKDAYS_SHORT_AM = ["እሑድ", "ሰኞ", "ማክ", "ረቡዕ", "ሐሙስ", "ዓርብ", "ቅዳሜ"];

  // Ge'ez Numerals (1 to 9999)
  function toGeezNumeral(num) {
    num = parseInt(num, 10);
    if (isNaN(num) || num <= 0 || num > 9999) return String(num);

    const digits = ["", "፩", "፪", "፫", "፬", "፭", "፮", "፯", "፰", "፱"];
    const tens = ["", "፲", "፳", "፴", "፵", "፶", "፷", "፸", "፹", "፺"];
    const hundreds = "፻";

    function under100(n) {
      const t = Math.floor(n / 10);
      const d = n % 10;
      return (t > 0 ? tens[t] : "") + (d > 0 ? digits[d] : "");
    }

    let result = "";
    const hPart = Math.floor(num / 100);
    const rem = num % 100;

    if (hPart > 0) {
      if (hPart > 1) result += under100(hPart);
      result += hundreds;
    }
    if (rem > 0) {
      result += under100(rem);
    }
    return result;
  }

  // Leap year check: year % 4 == 3
  function isEthiopianLeapYear(year) {
    return (year % 4) === 3;
  }

  function daysInEthiopianMonth(year, month) {
    if (month < 1 || month > 13) return 0;
    if (month <= 12) return 30;
    return isEthiopianLeapYear(year) ? 6 : 5;
  }

  // Mathematical Bijective Julian Day Number (JDN) Conversions
  function gregorianToJDN(year, month, day) {
    const a = Math.floor((14 - month) / 12);
    const y = year + 4800 - a;
    const m = month + 12 * a - 3;
    return day + Math.floor((153 * m + 2) / 5) + 365 * y + Math.floor(y / 4) - Math.floor(y / 100) + Math.floor(y / 400) - 32045;
  }

  function jdnToEthiopian(jdn) {
    const offset = jdn - ETHIOPIC_JDN_EPOCH;
    let y = Math.floor((4 * offset + 1463) / 1461);
    let priorDays = 365 * (y - 1) + Math.floor(y / 4);
    let dayInYear = offset - priorDays;

    if (dayInYear < 0) {
      y--;
      priorDays = 365 * (y - 1) + Math.floor(y / 4);
      dayInYear = offset - priorDays;
    }

    const month = Math.floor(dayInYear / 30) + 1;
    const day = (dayInYear % 30) + 1;
    return { year: y, month, day };
  }

  function ethiopianToJDN(year, month, day) {
    return ETHIOPIC_JDN_EPOCH + 365 * (year - 1) + Math.floor(year / 4) + (month - 1) * 30 + (day - 1);
  }

  function jdnToGregorian(jdn) {
    const a = jdn + 32044;
    const b = Math.floor((4 * a + 3) / 146097);
    const c = a - Math.floor((146097 * b) / 4);
    const d = Math.floor((4 * c + 3) / 1461);
    const e = c - Math.floor((1461 * d) / 4);
    const m = Math.floor((5 * e + 2) / 153);
    const day = e - Math.floor((153 * m + 2) / 5) + 1;
    const month = m + 3 - 12 * Math.floor(m / 10);
    const year = 100 * b + d - 4800 + Math.floor(m / 10);
    return { year, month, day };
  }

  // Convert Date object / ISO string to EthiopianDate
  function toEthiopian(dateInput) {
    const d = (dateInput instanceof Date) ? dateInput : new Date(dateInput);
    if (isNaN(d.getTime())) return null;

    // Convert in Africa/Addis_Ababa timezone (+3 UTC)
    const utcMs = d.getTime() + (d.getTimezoneOffset() * 60000);
    const addisMs = utcMs + (3 * 3600000);
    const addisDate = new Date(addisMs);

    const gy = addisDate.getFullYear();
    const gm = addisDate.getMonth() + 1;
    const gd = addisDate.getDate();

    const jdn = gregorianToJDN(gy, gm, gd);
    return jdnToEthiopian(jdn);
  }

  // Convert Ethiopian Date to Gregorian Date
  function toGregorian(year, month, day) {
    const jdn = ethiopianToJDN(year, month, day);
    const g = jdnToGregorian(jdn);
    return new Date(Date.UTC(g.year, g.month - 1, g.day, 0, 0, 0));
  }

  // Format Ethiopian Date string
  function formatEthiopianString(ethDate, lang, useGeez) {
    if (!ethDate) return '';
    const mName = (lang === 'am') ? ETHIOPIAN_MONTHS_AM[ethDate.month] : ETHIOPIAN_MONTHS_EN[ethDate.month];
    const dayStr = useGeez ? toGeezNumeral(ethDate.day) : ethDate.day;
    const yearStr = useGeez ? toGeezNumeral(ethDate.year) : ethDate.year;
    return `${mName} ${dayStr}, ${yearStr}`;
  }

  // Global Settings for Calendar Presentation
  const SETTING_CALENDAR_MODE_KEY = 'civic_calendar_mode'; // 'auto', 'ethiopian', 'gregorian'
  const SETTING_GEEZ_NUMERALS_KEY = 'civic_geez_numerals'; // 'true', 'false'

  function getCalendarDisplayMode() {
    return localStorage.getItem(SETTING_CALENDAR_MODE_KEY) || 'auto';
  }

  function setCalendarDisplayMode(mode) {
    localStorage.setItem(SETTING_CALENDAR_MODE_KEY, mode);
  }

  function isGeezNumeralsEnabled() {
    return localStorage.getItem(SETTING_GEEZ_NUMERALS_KEY) === 'true';
  }

  function setGeezNumeralsEnabled(enabled) {
    localStorage.setItem(SETTING_GEEZ_NUMERALS_KEY, enabled ? 'true' : 'false');
  }

  // Civic Date Formatter respecting Language toggle & Settings
  // In Amharic: shows Ethiopian date ("መስከረም 17, 2019" or Ge'ez "መስከረም ፲፯, ፳፻፲፱")
  // In English: shows Gregorian date with Ethiopian date in brackets ("Sep 27, 2026 [Meskerem 17, 2019]")
  function formatCivicDate(dateInput, options = {}) {
    if (!dateInput) return '--';
    const d = (dateInput instanceof Date) ? dateInput : new Date(dateInput);
    if (isNaN(d.getTime())) return String(dateInput);

    const lang = options.lang || (window.currentLang || 'en');
    const mode = options.mode || getCalendarDisplayMode();
    const useGeez = (options.useGeez !== undefined) ? options.useGeez : isGeezNumeralsEnabled();
    const showTime = options.showTime !== false;

    const eth = toEthiopian(d);
    const ethStr = eth ? formatEthiopianString(eth, lang, useGeez) : '';
    const ethStrEn = eth ? formatEthiopianString(eth, 'en', false) : '';

    // Gregorian representation
    const gy = d.getFullYear();
    const gm = d.getMonth();
    const gd = d.getDate();
    const mNameGreg = (lang === 'am') ? GREGORIAN_MONTHS_AM[gm] : GREGORIAN_MONTHS_EN[gm];
    const gregDateStr = `${mNameGreg} ${gd}, ${gy}`;

    // Time string (e.g. 14:30)
    let timeStr = '';
    if (showTime) {
      const hh = String(d.getHours()).padStart(2, '0');
      const mm = String(d.getMinutes()).padStart(2, '0');
      timeStr = ` ${hh}:${mm}`;
    }

    if (mode === 'ethiopian') {
      return ethStr + timeStr;
    }
    if (mode === 'gregorian') {
      return gregDateStr + timeStr;
    }

    // Default 'auto': follows language toggle
    if (lang === 'am') {
      return ethStr + timeStr;
    } else {
      // English: "Sep 27, 2026 [Meskerem 17, 2019]"
      return `${gregDateStr}${timeStr} [${ethStrEn}]`;
    }
  }

  // Public Holiday Cache
  let cachedHolidays = [];
  let holidaysMap = {}; // "YYYY-MM-DD" -> Holiday Object

  async function loadHolidays(year = 0) {
    try {
      const q = year > 0 ? `?year=${year}` : '';
      const endpoint = (typeof window.apiUrl === 'function') ? window.apiUrl(`/api/v1/calendar/holidays${q}`) : `/api/v1/calendar/holidays${q}`;
      const res = await fetch(endpoint);
      if (!res.ok) return [];

      const data = await res.json();
      const list = data.holidays || data || [];
      cachedHolidays = list;
      holidaysMap = {};
      list.forEach(h => {
        if (h.active) {
          const key = h.date.split('T')[0];
          holidaysMap[key] = h;
        }
      });
      return list;
    } catch (e) {
      console.warn("Could not fetch holidays:", e);
      return [];
    }
  }

  function getHolidayForDate(dateStr) {
    if (!dateStr) return null;
    const key = (dateStr instanceof Date) ? dateStr.toISOString().split('T')[0] : dateStr.split('T')[0];
    return holidaysMap[key] || null;
  }

  // Working Days Left Badge Component
  // Green, Amber, Red states with rich tooltip showing deadline and holiday skips
  function renderWorkingDaysBadge(caseItem) {
    if (!caseItem) return '';
    const status = caseItem.status;
    const isBreached = caseItem.is_breached;
    const isResolved = status === 'RESOLVED' || status === 'REJECTED';

    if (isResolved) {
      return `<span class="sla-work-badge badge-resolved" title="Case resolved on schedule">✓ ${window.currentLang === 'am' ? 'የተጠናቀቀ' : 'Resolved'}</span>`;
    }

    const workingDays = (caseItem.working_days_remaining !== undefined && caseItem.working_days_remaining !== null)
      ? parseFloat(caseItem.working_days_remaining)
      : Math.max(0, parseFloat((caseItem.remaining_hours || 0) / 8.5));

    const deadlineFormatted = formatCivicDate(caseItem.sla_deadline, { showTime: true });
    const tooltip = `SLA Deadline: ${deadlineFormatted} • Office Hours: 08:30 - 17:00 (UTC+3) • Skips weekends & official public holidays`;

    let stateClass = 'badge-green';
    let icon = '🟢';
    if (isBreached || workingDays <= 0) {
      stateClass = 'badge-red';
      icon = '🚨';
    } else if (workingDays <= 1.0) {
      stateClass = 'badge-red';
      icon = '🔴';
    } else if (workingDays <= 2.0) {
      stateClass = 'badge-amber';
      icon = '🟡';
    }

    let text = '';
    if (isBreached || workingDays <= 0) {
      text = window.currentLang === 'am' ? `${icon} ጊዜው አልፏል (Breached)` : `${icon} Overdue (Breached)`;
    } else {
      const roundedDays = workingDays.toFixed(1);
      text = window.currentLang === 'am' ? `${icon} ${roundedDays} የስራ ቀናት ይቀራሉ` : `${icon} ${roundedDays} Working Days Left`;
    }

    return `<span class="sla-work-badge ${stateClass}" title="${tooltip}" data-tooltip="${tooltip}">${text}</span>`;
  }

  // ============================================================================
  // 13-MONTH ETHIOPIAN DATE PICKER COMPONENT
  // ============================================================================
  let activePickerTarget = null;
  let activePickerOptions = {};
  let currentPickerEthDate = null; // { year, month, day }

  function createDatePickerModal() {
    if (document.getElementById('ethDatePickerModal')) return;

    const modal = document.createElement('div');
    modal.id = 'ethDatePickerModal';
    modal.className = 'modal-backdrop hidden';
    modal.onclick = function(e) {
      if (e.target === modal) closeEthiopianDatePicker();
    };

    modal.innerHTML = `
      <div class="modal-dialog eth-datepicker-dialog">
        <div class="modal-header eth-datepicker-header">
          <div style="display:flex; align-items:center; gap:0.6rem;">
            <span style="font-size:1.4rem;">📅</span>
            <div>
              <h3 class="modal-title" id="ethPickerTitle">የኢትዮጵያ ቀን መቁጠሪያ (13 ወራት)</h3>
              <p class="modal-subtitle" id="ethPickerSubtitle">Ethiopian Calendar Date Selector</p>
            </div>
          </div>
          <button class="modal-close" onclick="closeEthiopianDatePicker()">&times;</button>
        </div>

        <div class="modal-body eth-datepicker-body">
          <!-- Calendar Header Navigation (Year & 13 Months) -->
          <div class="eth-picker-controls">
            <div class="eth-year-stepper">
              <button type="button" class="btn btn-secondary btn-sm" onclick="changePickerYear(-1)">◀</button>
              <span id="ethPickerYearDisplay" class="eth-year-display">2019 ዓ.ም (2026/27)</span>
              <button type="button" class="btn btn-secondary btn-sm" onclick="changePickerYear(1)">▶</button>
            </div>

            <!-- Month Quick Selector Tabs (1 to 13) -->
            <div class="eth-months-strip" id="ethPickerMonthsStrip">
              <!-- Injected dynamically -->
            </div>
          </div>

          <!-- Notice banner for Hearing Scheduler Mode -->
          <div id="ethPickerHearingNotice" class="hearing-rule-note" style="display:none; margin:0.75rem 0;">
            <span class="note-icon">⚖️</span>
            <span><strong>Hearing Mode Active:</strong> Only valid <strong>Wednesday</strong> and <strong>Friday</strong> non-holiday dispute desk sessions are selectable.</span>
          </div>

          <!-- Day of Week Headers -->
          <div class="eth-grid-weekdays">
            <span>ሰኞ<br><small>Mon</small></span>
            <span>ማክሰኞ<br><small>Tue</small></span>
            <span class="hearing-col-header">ረቡዕ ⭐<br><small>Wed</small></span>
            <span>ሐሙስ<br><small>Thu</small></span>
            <span class="hearing-col-header">ዓርብ ⭐<br><small>Fri</small></span>
            <span class="weekend-col-header">ቅዳሜ<br><small>Sat</small></span>
            <span class="weekend-col-header">እሑድ<br><small>Sun</small></span>
          </div>

          <!-- Days Grid (1 to 30 or Pagume 5/6) -->
          <div class="eth-grid-days" id="ethPickerDaysGrid">
            <!-- Injected dynamically -->
          </div>

          <!-- Selected Date Display Strip -->
          <div class="eth-selected-preview" id="ethPickerSelectedPreview">
            <span>Selected: None</span>
          </div>
        </div>

        <div class="modal-footer eth-datepicker-footer">
          <div style="display:flex; gap:0.5rem; align-items:center;">
            <span class="legend-dot holiday-dot"></span><small>Holiday (የመንግስት በዓል)</small>
            <span class="legend-dot hearing-dot"></span><small>Hearing Day (ረቡዕ/ዓርብ)</small>
          </div>
          <div style="display:flex; gap:0.5rem;">
            <button type="button" class="btn btn-secondary btn-sm" onclick="selectPickerToday()">Today (ዛሬ)</button>
            <button type="button" class="btn btn-primary btn-sm" onclick="confirmPickerSelection()">Confirm Selection</button>
          </div>
        </div>
      </div>
    `;

    document.body.appendChild(modal);
  }

  function openEthiopianDatePicker(targetInput, options = {}) {
    createDatePickerModal();
    activePickerTarget = targetInput;
    activePickerOptions = options;

    const notice = document.getElementById('ethPickerHearingNotice');
    if (notice) notice.style.display = options.hearingOnly ? 'flex' : 'none';

    // Determine starting date
    let startGreg = new Date();
    if (targetInput && targetInput.value) {
      const parsed = new Date(targetInput.value);
      if (!isNaN(parsed.getTime())) startGreg = parsed;
    }
    currentPickerEthDate = toEthiopian(startGreg) || { year: 2019, month: 1, day: 1 };

    renderPickerMonthStrip();
    renderPickerDays();

    const modal = document.getElementById('ethDatePickerModal');
    if (modal) modal.classList.remove('hidden');
  }

  function closeEthiopianDatePicker() {
    const modal = document.getElementById('ethDatePickerModal');
    if (modal) modal.classList.add('hidden');
    activePickerTarget = null;
  }

  function changePickerYear(delta) {
    if (!currentPickerEthDate) return;
    currentPickerEthDate.year += delta;
    if (currentPickerEthDate.year < 2000) currentPickerEthDate.year = 2000;
    if (currentPickerEthDate.year > 2050) currentPickerEthDate.year = 2050;
    renderPickerDays();
  }

  function setPickerMonth(month) {
    if (!currentPickerEthDate) return;
    currentPickerEthDate.month = month;
    const maxDays = daysInEthiopianMonth(currentPickerEthDate.year, month);
    if (currentPickerEthDate.day > maxDays) currentPickerEthDate.day = maxDays;
    renderPickerMonthStrip();
    renderPickerDays();
  }

  function renderPickerMonthStrip() {
    const strip = document.getElementById('ethPickerMonthsStrip');
    if (!strip || !currentPickerEthDate) return;

    strip.innerHTML = '';
    for (let m = 1; m <= 13; m++) {
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.className = `eth-month-pill ${m === currentPickerEthDate.month ? 'active' : ''} ${m === 13 ? 'pagume-pill' : ''}`;
      const amName = ETHIOPIAN_MONTHS_AM[m];
      const enName = ETHIOPIAN_MONTHS_EN[m];
      btn.innerHTML = `<span>${amName}</span><small>${enName}</small>`;
      btn.onclick = () => setPickerMonth(m);
      strip.appendChild(btn);
    }
  }

  function renderPickerDays() {
    const grid = document.getElementById('ethPickerDaysGrid');
    const yearDisplay = document.getElementById('ethPickerYearDisplay');
    const preview = document.getElementById('ethPickerSelectedPreview');
    if (!grid || !currentPickerEthDate) return;

    const y = currentPickerEthDate.year;
    const m = currentPickerEthDate.month;
    const d = currentPickerEthDate.day;

    const greg1st = toGregorian(y, m, 1);
    const gregYear = greg1st.getFullYear();
    if (yearDisplay) {
      yearDisplay.textContent = `${y} ዓ.ም (${gregYear}/${gregYear + 1})`;
    }

    grid.innerHTML = '';

    // Day of week of month 1st (0 = Sun, 1 = Mon ... 6 = Sat)
    // Align grid so 0 is Monday: (day + 6) % 7
    let firstWd = greg1st.getUTCDay();
    let padDays = (firstWd + 6) % 7;

    for (let p = 0; p < padDays; p++) {
      const padCell = document.createElement('div');
      padCell.className = 'eth-day-cell empty-pad';
      grid.appendChild(padCell);
    }

    const totalDays = daysInEthiopianMonth(y, m);

    for (let dayNum = 1; dayNum <= totalDays; dayNum++) {
      const dayGreg = toGregorian(y, m, dayNum);
      const gregIso = dayGreg.toISOString().split('T')[0];
      const weekday = dayGreg.getUTCDay(); // 0 = Sun, 3 = Wed, 5 = Fri, 6 = Sat
      const isWeekend = (weekday === 0 || weekday === 6);
      const isHearingDay = (weekday === 3 || weekday === 5);
      const holiday = getHolidayForDate(gregIso);
      const isSelected = (dayNum === d);

      const cell = document.createElement('button');
      cell.type = 'button';
      cell.className = 'eth-day-cell';

      if (isWeekend) cell.classList.add('weekend-day');
      if (holiday) cell.classList.add('holiday-day');
      if (isHearingDay && !holiday) cell.classList.add('hearing-day');
      if (isSelected) cell.classList.add('selected-day');

      // Hearing Only Restriction
      const isDisabledHearing = activePickerOptions.hearingOnly && (!isHearingDay || holiday);
      if (isDisabledHearing) {
        cell.classList.add('disabled-hearing');
        cell.disabled = true;
      }

      // Cell markup
      const geezDay = isGeezNumeralsEnabled() ? toGeezNumeral(dayNum) : dayNum;
      const gregSmall = `${dayGreg.getUTCMonth() + 1}/${dayGreg.getUTCDate()}`;

      let badgeHtml = '';
      if (holiday) {
        badgeHtml = `<span class="cell-holiday-tag" title="${holiday.name_am} (${holiday.name_en})">🎉 ${holiday.name_am}</span>`;
      } else if (isHearingDay && activePickerOptions.hearingOnly) {
        badgeHtml = `<span class="cell-hearing-tag">⚖️ Slot</span>`;
      }

      cell.innerHTML = `
        <span class="day-number">${geezDay}</span>
        <span class="day-greg-sub">${gregSmall}</span>
        ${badgeHtml}
      `;

      if (!isDisabledHearing) {
        cell.onclick = () => {
          currentPickerEthDate.day = dayNum;
          renderPickerDays();
        };
      }

      grid.appendChild(cell);
    }

    // Update Preview
    if (preview) {
      const selGreg = toGregorian(y, m, d);
      const selIso = selGreg.toISOString().split('T')[0];
      const selHoliday = getHolidayForDate(selIso);
      const holNotice = selHoliday ? ` • <strong>🎉 ${selHoliday.name_am} (${selHoliday.name_en})</strong>` : '';
      const ethLabel = formatEthiopianString(currentPickerEthDate, 'am', isGeezNumeralsEnabled());
      const gregLabel = selIso;

      preview.innerHTML = `
        <div>
          <span>Selected Date:</span>
          <strong>${ethLabel}</strong> (${gregLabel}) ${holNotice}
        </div>
      `;
    }
  }

  function selectPickerToday() {
    const today = new Date();
    currentPickerEthDate = toEthiopian(today) || { year: 2019, month: 1, day: 1 };
    renderPickerMonthStrip();
    renderPickerDays();
  }

  function confirmPickerSelection() {
    if (!activePickerTarget || !currentPickerEthDate) {
      closeEthiopianDatePicker();
      return;
    }

    const selGreg = toGregorian(currentPickerEthDate.year, currentPickerEthDate.month, currentPickerEthDate.day);
    const isoDate = selGreg.toISOString().split('T')[0];
    const ethLabel = formatEthiopianString(currentPickerEthDate, window.currentLang || 'en', isGeezNumeralsEnabled());

    // Update target input
    activePickerTarget.value = isoDate;
    activePickerTarget.setAttribute('data-ethiopian', ethLabel);

    // If there's an associated display label element, update it
    const displayLabelId = activePickerTarget.getAttribute('data-display-label');
    if (displayLabelId) {
      const labelEl = document.getElementById(displayLabelId);
      if (labelEl) labelEl.textContent = `${ethLabel} (${isoDate})`;
    }

    // Trigger change event
    activePickerTarget.dispatchEvent(new Event('change', { bubbles: true }));

    if (typeof activePickerOptions.onSelect === 'function') {
      activePickerOptions.onSelect(isoDate, currentPickerEthDate, ethLabel);
    }

    closeEthiopianDatePicker();
  }

  // ============================================================================
  // OFFICIAL CALENDAR PAGE VIEW & WORKING-DAY CALCULATOR TOOL
  // ============================================================================
  let calPageCurrentYear = 2026;
  let calPageCurrentMonth = 10; // Gregorian Oct 2026 / Meskerem-Tikimt 2019
  let calPageMode = 'ethiopian'; // 'ethiopian' or 'gregorian'

  async function initCalendarPage() {
    await loadHolidays(calPageCurrentYear);
    renderCalendarPageGrid();
    renderHolidaysDirectoryTable();
  }

  function setCalendarPageMode(mode) {
    calPageMode = mode;
    const btnEth = document.getElementById('calModeBtnEth');
    const btnGreg = document.getElementById('calModeBtnGreg');
    if (btnEth) btnEth.classList.toggle('active', mode === 'ethiopian');
    if (btnGreg) btnGreg.classList.toggle('active', mode === 'gregorian');
    renderCalendarPageGrid();
  }

  function changeCalendarPageMonth(delta) {
    if (calPageMode === 'ethiopian') {
      if (!window._calPageEthMonth) window._calPageEthMonth = 1;
      if (!window._calPageEthYear) window._calPageEthYear = 2019;
      window._calPageEthMonth += delta;
      if (window._calPageEthMonth > 13) {
        window._calPageEthMonth = 1;
        window._calPageEthYear++;
      } else if (window._calPageEthMonth < 1) {
        window._calPageEthMonth = 13;
        window._calPageEthYear--;
      }
    } else {
      calPageCurrentMonth += delta;
      if (calPageCurrentMonth > 12) {
        calPageCurrentMonth = 1;
        calPageCurrentYear++;
      } else if (calPageCurrentMonth < 1) {
        calPageCurrentMonth = 12;
        calPageCurrentYear--;
      }
    }
    renderCalendarPageGrid();
  }

  function renderCalendarPageGrid() {
    const grid = document.getElementById('civicCalendarMonthGrid');
    const titleEl = document.getElementById('civicCalendarMonthHeading');
    if (!grid) return;

    grid.innerHTML = '';

    if (calPageMode === 'ethiopian') {
      if (!window._calPageEthMonth) window._calPageEthMonth = 1;
      if (!window._calPageEthYear) window._calPageEthYear = 2019;
      const ey = window._calPageEthYear;
      const em = window._calPageEthMonth;

      const mNameAm = ETHIOPIAN_MONTHS_AM[em];
      const mNameEn = ETHIOPIAN_MONTHS_EN[em];
      if (titleEl) {
        titleEl.textContent = `${mNameAm} (${mNameEn}) ${ey} ዓ.ም`;
      }

      const greg1st = toGregorian(ey, em, 1);
      const firstWd = greg1st.getUTCDay();
      const padDays = (firstWd + 6) % 7;

      for (let p = 0; p < padDays; p++) {
        const pad = document.createElement('div');
        pad.className = 'civic-cal-cell empty-pad';
        grid.appendChild(pad);
      }

      const totalDays = daysInEthiopianMonth(ey, em);
      for (let d = 1; d <= totalDays; d++) {
        const dayGreg = toGregorian(ey, em, d);
        const iso = dayGreg.toISOString().split('T')[0];
        const wd = dayGreg.getUTCDay();
        const isWeekend = (wd === 0 || wd === 6);
        const isHearing = (wd === 3 || wd === 5);
        const holiday = getHolidayForDate(iso);

        const cell = document.createElement('div');
        cell.className = 'civic-cal-cell';
        if (isWeekend) cell.classList.add('weekend-cell');
        if (holiday) cell.classList.add('holiday-cell');
        if (isHearing && !holiday) cell.classList.add('hearing-cell');

        let tags = '';
        if (holiday) {
          tags += `<div class="cell-event-badge holiday-badge" title="${holiday.source}">🎉 ${holiday.name_am} (${holiday.type})</div>`;
        }
        if (isHearing && !holiday) {
          tags += `<div class="cell-event-badge hearing-badge">⚖️ Hearing Desk Session</div>`;
        }

        cell.innerHTML = `
          <div class="cell-date-bar">
            <span class="eth-day-bold">${isGeezNumeralsEnabled() ? toGeezNumeral(d) : d}</span>
            <span class="greg-day-mute">${dayGreg.getUTCMonth() + 1}/${dayGreg.getUTCDate()}</span>
          </div>
          <div class="cell-events-tray">
            ${tags}
          </div>
        `;
        grid.appendChild(cell);
      }
    } else {
      // Gregorian view
      const gy = calPageCurrentYear;
      const gm = calPageCurrentMonth;
      const mName = GREGORIAN_MONTHS_EN[gm - 1];
      if (titleEl) {
        titleEl.textContent = `${mName} ${gy}`;
      }

      const firstDate = new Date(Date.UTC(gy, gm - 1, 1));
      const padDays = (firstDate.getUTCDay() + 6) % 7;
      for (let p = 0; p < padDays; p++) {
        const pad = document.createElement('div');
        pad.className = 'civic-cal-cell empty-pad';
        grid.appendChild(pad);
      }

      const lastDay = new Date(Date.UTC(gy, gm, 0)).getUTCDate();
      for (let d = 1; d <= lastDay; d++) {
        const dObj = new Date(Date.UTC(gy, gm - 1, d));
        const iso = dObj.toISOString().split('T')[0];
        const eth = toEthiopian(dObj);
        const wd = dObj.getUTCDay();
        const isWeekend = (wd === 0 || wd === 6);
        const isHearing = (wd === 3 || wd === 5);
        const holiday = getHolidayForDate(iso);

        const cell = document.createElement('div');
        cell.className = 'civic-cal-cell';
        if (isWeekend) cell.classList.add('weekend-cell');
        if (holiday) cell.classList.add('holiday-cell');
        if (isHearing && !holiday) cell.classList.add('hearing-cell');

        let tags = '';
        if (holiday) {
          tags += `<div class="cell-event-badge holiday-badge">🎉 ${holiday.name_am}</div>`;
        }
        if (isHearing && !holiday) {
          tags += `<div class="cell-event-badge hearing-badge">⚖️ Hearing Desk</div>`;
        }

        cell.innerHTML = `
          <div class="cell-date-bar">
            <span class="eth-day-bold">${d}</span>
            <span class="greg-day-mute">${eth ? ETHIOPIAN_MONTHS_AM[eth.month] + ' ' + eth.day : ''}</span>
          </div>
          <div class="cell-events-tray">
            ${tags}
          </div>
        `;
        grid.appendChild(cell);
      }
    }
  }

  // Render Holidays Directory Table
  function renderHolidaysDirectoryTable() {
    const tbody = document.getElementById('officialHolidaysTableBody');
    if (!tbody) return;

    tbody.innerHTML = '';
    const isSuperAdmin = window.currentUser && window.currentUser.role === 'SUPER_ADMIN';

    cachedHolidays.forEach(h => {
      const tr = document.createElement('tr');
      const dObj = new Date(h.date);
      const eth = toEthiopian(dObj);
      const ethFormatted = eth ? formatEthiopianString(eth, 'am', isGeezNumeralsEnabled()) : h.ethiopian_date_am;

      let actionsHtml = isSuperAdmin ? `
        <div style="display:flex; gap:0.4rem;">
          <button type="button" class="btn btn-secondary btn-sm" onclick="CivicCalendar.openEditHolidayModal(${h.id})" title="Edit Moon Sighting or Details">✏️ Edit</button>
          <button type="button" class="btn btn-danger btn-sm" onclick="CivicCalendar.deleteHoliday(${h.id})" title="Delete Holiday">🗑️</button>
        </div>
      ` : '<span style="color:#94a3b8; font-size:0.75rem;">Official Gazetted</span>';

      tr.innerHTML = `
        <td><strong>${h.date.split('T')[0]}</strong></td>
        <td><span class="eth-date-tag">${ethFormatted}</span></td>
        <td><div><strong>${h.name_am}</strong></div><div style="font-size:0.75rem; color:#94a3b8;">${h.name_en}</div></td>
        <td><span class="holiday-type-badge type-${h.type}">${h.type}</span></td>
        <td><span style="font-size:0.75rem; color:#cbd5e1;">${h.source || 'Civil Proclamation No. 16/1975'}</span></td>
        <td>${actionsHtml}</td>
      `;
      tbody.appendChild(tr);
    });
  }

  // Working-Day Counter Tool ("from X to Y = N working days")
  async function calculateWorkingDaysTool() {
    const fromInput = document.getElementById('calcFromInput');
    const toInput = document.getElementById('calcToInput');
    const calSelect = document.getElementById('calcCalendarSelect');
    const resultBox = document.getElementById('calcResultsDisplay');

    if (!fromInput || !toInput || !resultBox) return;

    const fromVal = fromInput.value.trim();
    const toVal = toInput.value.trim();
    const calVal = calSelect ? calSelect.value : 'gregorian';

    if (!fromVal || !toVal) {
      if (typeof window.showToast === 'function') {
        window.showToast('Please provide both Start and End dates', 'warning');
      }
      return;
    }

    resultBox.innerHTML = '<div style="padding:1rem; text-align:center; color:var(--cyan-400);">⏳ Calculating working business days...</div>';

    try {
      const q = `?from=${encodeURIComponent(fromVal)}&to=${encodeURIComponent(toVal)}&calendar=${encodeURIComponent(calVal)}`;
      const endpoint = (typeof window.apiUrl === 'function') ? window.apiUrl(`/api/v1/calendar/working-days${q}`) : `/api/v1/calendar/working-days${q}`;
      const res = await fetch(endpoint);
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Calculation failed');
      }

      const data = await res.json();
      resultBox.innerHTML = `
        <div class="calc-results-card">
          <div class="calc-hero-metric">
            <span class="calc-num">${data.working_days.toFixed(1)}</span>
            <span class="calc-unit">${window.currentLang === 'am' ? 'የስራ ቀናት (Business Days)' : 'Working Days'}</span>
          </div>

          <div class="calc-stat-grid">
            <div class="calc-stat-item">
              <span class="calc-stat-label">From:</span>
              <strong>${data.from_ethiopian_am}</strong>
              <small>(${data.from_gregorian})</small>
            </div>
            <div class="calc-stat-item">
              <span class="calc-stat-label">To:</span>
              <strong>${data.to_ethiopian_am}</strong>
              <small>(${data.to_gregorian})</small>
            </div>
            <div class="calc-stat-item">
              <span class="calc-stat-label">Calendar Days:</span>
              <strong>${data.total_calendar_days} Days</strong>
            </div>
            <div class="calc-stat-item">
              <span class="calc-stat-label">Holidays Skipped:</span>
              <strong style="color:var(--warm-gold);">${data.holidays_encountered} Days</strong>
            </div>
            <div class="calc-stat-item">
              <span class="calc-stat-label">Weekends Skipped:</span>
              <strong style="color:#94a3b8;">${data.weekend_days_skipped} Days</strong>
            </div>
            <div class="calc-stat-item">
              <span class="calc-stat-label">Office Schedule:</span>
              <strong>08:30 - 17:00 (8.5h/d)</strong>
            </div>
          </div>
        </div>
      `;
    } catch (e) {
      resultBox.innerHTML = `<div style="padding:1rem; color:#ef4444;">Error: ${e.message}</div>`;
    }
  }

  // Super Admin Holiday Management Handlers
  function openAddHolidayModal() {
    const modal = document.getElementById('holidayAdminModal');
    if (!modal) return;
    document.getElementById('holidayModalTitle').textContent = 'Add Public Holiday (የህዝብ በዓል መዝግብ)';
    document.getElementById('holidayEditId').value = '';
    document.getElementById('holidayDateInput').value = '';
    document.getElementById('holidayNameEnInput').value = '';
    document.getElementById('holidayNameAmInput').value = '';
    document.getElementById('holidayTypeSelect').value = 'FIXED';
    document.getElementById('holidaySourceInput').value = 'Federal Civil Service & Proclamation';
    modal.classList.remove('hidden');
  }

  function openEditHolidayModal(id) {
    const h = cachedHolidays.find(item => item.id === id);
    if (!h) return;
    const modal = document.getElementById('holidayAdminModal');
    if (!modal) return;
    document.getElementById('holidayModalTitle').textContent = `Edit Holiday #${h.id} (${h.name_en})`;
    document.getElementById('holidayEditId').value = h.id;
    document.getElementById('holidayDateInput').value = h.date.split('T')[0];
    document.getElementById('holidayNameEnInput').value = h.name_en;
    document.getElementById('holidayNameAmInput').value = h.name_am;
    document.getElementById('holidayTypeSelect').value = h.type;
    document.getElementById('holidaySourceInput').value = h.source || '';
    modal.classList.remove('hidden');
  }

  async function handleHolidayAdminSubmit(e) {
    e.preventDefault();
    const id = document.getElementById('holidayEditId').value;
    const date = document.getElementById('holidayDateInput').value.trim();
    const name_en = document.getElementById('holidayNameEnInput').value.trim();
    const name_am = document.getElementById('holidayNameAmInput').value.trim();
    const type = document.getElementById('holidayTypeSelect').value;
    const source = document.getElementById('holidaySourceInput').value.trim();

    const isEdit = !!id;
    const method = isEdit ? 'PUT' : 'POST';
    const path = isEdit ? `/api/v1/calendar/holidays/${id}` : '/api/v1/calendar/holidays';
    const endpoint = (typeof window.apiUrl === 'function') ? window.apiUrl(path) : path;

    try {
      const res = await fetch(endpoint, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + (window.currentToken || '')
        },
        body: JSON.stringify({ date, name_en, name_am, type, source })
      });

      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Failed to save holiday');
      }

      if (typeof window.showToast === 'function') {
        window.showToast(isEdit ? 'Holiday updated successfully!' : 'New holiday registered!', 'success');
      }
      closeModal('holidayAdminModal');
      await initCalendarPage();
    } catch (err) {
      if (typeof window.showToast === 'function') {
        window.showToast(err.message, 'danger');
      }
    }
  }

  async function deleteHoliday(id) {
    if (!confirm('Are you sure you want to delete this public holiday from the official calendar?')) return;
    const endpoint = (typeof window.apiUrl === 'function') ? window.apiUrl(`/api/v1/calendar/holidays/${id}`) : `/api/v1/calendar/holidays/${id}`;

    try {
      const res = await fetch(endpoint, {
        method: 'DELETE',
        headers: {
          'Authorization': 'Bearer ' + (window.currentToken || '')
        }
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Failed to delete');
      }
      if (typeof window.showToast === 'function') {
        window.showToast('Holiday removed', 'info');
      }
      await initCalendarPage();
    } catch (e) {
      if (typeof window.showToast === 'function') {
        window.showToast(e.message, 'danger');
      }
    }
  }

  // Export to Global namespace
  window.CivicCalendar = {
    toEthiopian,
    toGregorian,
    toGeezNumeral,
    isEthiopianLeapYear,
    daysInEthiopianMonth,
    formatCivicDate,
    renderWorkingDaysBadge,
    getCalendarDisplayMode,
    setCalendarDisplayMode,
    isGeezNumeralsEnabled,
    setGeezNumeralsEnabled,
    loadHolidays,
    getHolidayForDate,
    openEthiopianDatePicker,
    closeEthiopianDatePicker,
    initCalendarPage,
    setCalendarPageMode,
    changeCalendarPageMonth,
    calculateWorkingDaysTool,
    openAddHolidayModal,
    openEditHolidayModal,
    handleHolidayAdminSubmit,
    deleteHoliday,
    ETHIOPIAN_MONTHS_EN,
    ETHIOPIAN_MONTHS_AM
  };

})(window);
