// Copyright (c) 2026 Heino Stömmer.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import './style.css';
import './app.css';

// Import Wails bindings for WailsAdapter
import {
    ScanLogs,
    GetAvailableMonths,
    GetSummaryForMonth,
    GetOverallTotal
} from '../wailsjs/go/ui/WailsAdapter';

// Internationalization (i18n) setup
const translations = {
    en: {
        subtitle: "Statistics of your local AI activities, credits and usage",
        btnRefresh: "Refresh",
        btnAbout: "About",
        loadingLogs: "Scanning VS Code Copilot logs...",
        selectMonthLabel: "Selected Month:",
        selectMonthLoading: "Loading months...",
        selectMonthNoData: "No data found",
        selectLangLabel: "Language:",
        overallTotal: "Overall Total:",
        overallTotalTokens: "Tokens",
        kpiTotalTokensTitle: "Monthly Total Tokens",
        kpiTotalTokensTrend: "Sum of all requests & completions",
        kpiPromptTokensTitle: "Prompt Tokens",
        kpiPromptTokensTrend: "Sent context",
        kpiCompletionTokensTitle: "Completion Tokens",
        kpiCompletionTokensTrend: "Generated code / responses",
        kpiTotalRequestsTitle: "Sent Prompts",
        kpiTotalRequestsTrend: "Number of interactions (Requests)",
        kpiTotalAicTitle: "AI Credits (AIC)",
        kpiTotalAicTrend: "Model-specific credits (AIC)",
        kpiTotalAiuTitle: "AI Usage (AIU)",
        kpiTotalAiuTrend: "Fine-grained compute (nano-AIU)",
        workspaceSectionTitle: "Consumption by VS Code Project (Workspace)",
        emptyStateNoData: "No data available. Please perform a scan.",
        emptyStateNoActivity: "No workspace activities in this month.",
        aboutTitle: "About this App",
        aboutBodyTitle: "Github Copilot Credit Count",
        aboutDescription: "This application analyzes local GitHub Copilot log files in VS Code to evaluate the tokens, prompts, completions used, as well as the model-specific AI Credits (AIC) and fine-grained compute units (AIU).",
        aboutLicense: "Licensed under the Apache License, Version 2.0",
        aboutClose: "Close",
        scanError: "Error scanning Copilot logs: ",
        wsPrompt: "Prompt",
        wsCompletion: "Completion",
        wsPrompts: "Prompts",
        wsCost: "Cost",
        wsUnknownPath: "Virtual/Unknown Workspace",
        startupTitle: "Analyzing Copilot Data...",
        startupDisclaimer: "Please note: The calculated metrics are only indicators and may be inaccurate or incomplete."
    },
    de: {
        subtitle: "Statistiken Ihrer lokalen KI-Aktivitäten, Credits und Usage",
        btnRefresh: "Aktualisieren",
        btnAbout: "Über",
        loadingLogs: "Scanne VS Code Copilot Protokolle...",
        selectMonthLabel: "Ausgewählter Monat:",
        selectMonthLoading: "Monate laden...",
        selectMonthNoData: "Keine Daten gefunden",
        selectLangLabel: "Sprache:",
        overallTotal: "Zähler-Gesamtstand:",
        overallTotalTokens: "Tokens",
        kpiTotalTokensTitle: "Monatliche Gesamt-Tokens",
        kpiTotalTokensTrend: "Summe aller Anfragen & Antworten",
        kpiPromptTokensTitle: "Prompt Tokens",
        kpiPromptTokensTrend: "Gesendeter Kontext",
        kpiCompletionTokensTitle: "Completion Tokens",
        kpiCompletionTokensTrend: "Generierter Code / Antworten",
        kpiTotalRequestsTitle: "Gesendete Prompts",
        kpiTotalRequestsTrend: "Anzahl der Interaktionen (Requests)",
        kpiTotalAicTitle: "AI Credits (AIC)",
        kpiTotalAicTrend: "Modellspezifische Credits (AIC)",
        kpiTotalAiuTitle: "AI Usage (AIU)",
        kpiTotalAiuTrend: "Feinstufige Rechenleistung (nano-AIU)",
        workspaceSectionTitle: "Verbrauch nach VS Code Projekt (Workspace)",
        emptyStateNoData: "Keine Daten vorhanden. Führen Sie einen Scan durch.",
        emptyStateNoActivity: "Keine Workspace-Aktivitäten in diesem Monat.",
        aboutTitle: "Über diese App",
        aboutBodyTitle: "Github Copilot Credit Count",
        aboutDescription: "Diese Anwendung analysiert die lokalen Protokolldateien von GitHub Copilot in VS Code, um die genutzten Tokens, Prompts, Completions und die modellspezifischen AI Credits (AIC) sowie die feingranularen Recheneinheiten (AIU) auszuwerten.",
        aboutLicense: "Lizenziert unter der Apache Lizenz, Version 2.0",
        aboutClose: "Schließen",
        scanError: "Fehler beim Scannen der Copilot-Protokolle: ",
        wsPrompt: "Prompt",
        wsCompletion: "Completion",
        wsPrompts: "Prompts",
        wsCost: "Kosten",
        wsUnknownPath: "Virtueller/Unbekannter Workspace",
        startupTitle: "Analysiere Copilot-Daten...",
        startupDisclaimer: "Hinweis: Die angezeigten Daten sind nur Anhaltspunkte und können fehlerhaft oder unvollständig sein."
    }
};

let currentLang = 'en';
let locale = 'en-US';

// State flags for startup logic
let isMinTimeElapsed = false;
let isInitialScanFinished = false;

// Retrieve translation by key
function t(key) {
    if (translations[currentLang] && translations[currentLang][key]) {
        return translations[currentLang][key];
    }
    if (translations['en'] && translations['en'][key]) {
        return translations['en'][key];
    }
    return key;
}

// Apply translations to static DOM elements
function applyStaticTranslations() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        const translation = t(key);
        if (translation !== key) {
            el.innerText = translation;
        }
    });
    document.documentElement.lang = currentLang;
}

// Cache elements
const monthSelect = document.getElementById('month-select');
const overallTotalEl = document.getElementById('overall-total');
const kpiTotalEl = document.getElementById('kpi-total-tokens');
const kpiPromptEl = document.getElementById('kpi-prompt-tokens');
const kpiCompletionEl = document.getElementById('kpi-completion-tokens');
const kpiRequestsEl = document.getElementById('kpi-total-requests');
const kpiAicEl = document.getElementById('kpi-total-aic');
const kpiAiuEl = document.getElementById('kpi-total-aiu');
const workspacesListEl = document.getElementById('workspaces-list');
const loadingOverlay = document.getElementById('loading-overlay');

let isScanning = false;

// Hide the startup warning banner once both conditions are met (scan complete & 5s elapsed)
function checkHideStartupBanner() {
    if (isMinTimeElapsed && isInitialScanFinished) {
        const banner = document.getElementById('startup-banner');
        if (banner) {
            banner.classList.add('hidden');
        }
    }
}

// Perform the initial background scan at startup
async function performInitialScan() {
    try {
        // Step 1: Scan Copilot logs in Go
        await ScanLogs();
        
        // Step 2: Fetch overall total token consumption
        const overallTotal = await GetOverallTotal();
        overallTotalEl.innerText = formatNumber(overallTotal.total) + ' ' + t('overallTotalTokens');
        
        // Step 3: Fetch available months list
        const months = await GetAvailableMonths();
        
        // Populate dropdown
        monthSelect.innerHTML = '';
        
        if (!months || months.length === 0) {
            monthSelect.innerHTML = `<option value="">${t('selectMonthNoData')}</option>`;
            clearDashboard();
            return;
        }
        
        months.forEach(m => {
            const opt = document.createElement('option');
            opt.value = m;
            opt.innerText = formatMonth(m);
            monthSelect.appendChild(opt);
        });
        
        monthSelect.value = months[0];
        await loadMonthData(months[0]);
        
    } catch (err) {
        console.error('Initial scan failed:', err);
        alert(t('scanError') + err);
    }
}

// Global language change function
window.onLangChanged = async function (newLang) {
    if (newLang === currentLang) return;
    currentLang = newLang;
    locale = currentLang === 'de' ? 'de-DE' : 'en-US';
    
    // Update active classes on language buttons
    document.querySelectorAll('.lang-btn').forEach(btn => {
        btn.classList.toggle('active', btn.getAttribute('data-lang') === currentLang);
    });
    
    // Translate all static texts
    applyStaticTranslations();
    
    // Save current selected month
    const currentMonth = monthSelect.value;
    
    try {
        loadingOverlay.classList.remove('hidden');
        
        // Reload overall total to refresh its text and number format
        const overallTotal = await GetOverallTotal();
        overallTotalEl.innerText = formatNumber(overallTotal.total) + ' ' + t('overallTotalTokens');
        
        // Reload available months
        const months = await GetAvailableMonths();
        monthSelect.innerHTML = '';
        
        if (!months || months.length === 0) {
            monthSelect.innerHTML = `<option value="">${t('selectMonthNoData')}</option>`;
            clearDashboard();
            return;
        }
        
        months.forEach(m => {
            const opt = document.createElement('option');
            opt.value = m;
            opt.innerText = formatMonth(m);
            monthSelect.appendChild(opt);
        });
        
        // Restore previous selected month or select the first one
        if (months.includes(currentMonth)) {
            monthSelect.value = currentMonth;
            await loadMonthData(currentMonth);
        } else {
            monthSelect.value = months[0];
            await loadMonthData(months[0]);
        }
    } catch (err) {
        console.error('Failed to change language:', err);
    } finally {
        loadingOverlay.classList.add('hidden');
    }
};

// Global trigger functions exposed to window (for manual refreshes)
window.triggerScan = async function () {
    if (isScanning) return;
    isScanning = true;
    
    // Show spinner
    loadingOverlay.classList.remove('hidden');
    
    try {
        // Step 1: Scan Copilot logs in Go
        await ScanLogs();
        
        // Step 2: Fetch overall total token consumption
        const overallTotal = await GetOverallTotal();
        overallTotalEl.innerText = formatNumber(overallTotal.total) + ' ' + t('overallTotalTokens');
        
        // Step 3: Fetch available months list
        const months = await GetAvailableMonths();
        
        // Save current selected month
        const prevSelected = monthSelect.value;
        
        // Populate dropdown
        monthSelect.innerHTML = '';
        
        if (!months || months.length === 0) {
            monthSelect.innerHTML = `<option value="">${t('selectMonthNoData')}</option>`;
            clearDashboard();
            return;
        }
        
        months.forEach(m => {
            const opt = document.createElement('option');
            opt.value = m;
            opt.innerText = formatMonth(m);
            monthSelect.appendChild(opt);
        });
        
        // Restore previous selection if still available, otherwise choose latest
        if (months.includes(prevSelected)) {
            monthSelect.value = prevSelected;
            await loadMonthData(prevSelected);
        } else {
            monthSelect.value = months[0];
            await loadMonthData(months[0]);
        }
        
    } catch (err) {
        console.error('Scan failed:', err);
        alert(t('scanError') + err);
    } finally {
        isScanning = false;
        loadingOverlay.classList.add('hidden');
    }
};

window.onMonthChanged = async function (selectedMonth) {
    if (!selectedMonth) return;
    loadingOverlay.classList.remove('hidden');
    try {
        await loadMonthData(selectedMonth);
    } catch (err) {
        console.error('Failed changing month:', err);
    } finally {
        loadingOverlay.classList.add('hidden');
    }
};

// Internal function to load and render monthly stats
async function loadMonthData(month) {
    const summary = await GetSummaryForMonth(month);
    
    // 1. Update KPI Cards
    kpiTotalEl.innerText = formatNumber(summary.totalTokens.total);
    kpiPromptEl.innerText = formatNumber(summary.totalTokens.prompt);
    kpiCompletionEl.innerText = formatNumber(summary.totalTokens.completion);
    kpiRequestsEl.innerText = formatNumber(summary.totalTokens.requests);
    kpiAicEl.innerText = formatCredit(summary.totalTokens.aic);
    kpiAiuEl.innerText = formatNumber(summary.totalTokens.aiu);
    
    // 2. Render Workspace list
    workspacesListEl.innerHTML = '';
    
    if (!summary.workspaces || summary.workspaces.length === 0) {
        workspacesListEl.innerHTML = `
            <div class="empty-state">
                <span class="empty-state-icon">📂</span>
                <p>${t('emptyStateNoActivity')}</p>
            </div>
        `;
        return;
    }
    
    summary.workspaces.forEach(wsSummary => {
        const totalMonthTokens = summary.totalTokens.total || 1;
        const percentage = Math.round((wsSummary.tokens.total / totalMonthTokens) * 100);
        
        const wsCard = document.createElement('div');
        wsCard.className = 'ws-item';
        
        // Decode path details if available
        const pathLabel = wsSummary.workspace.path || t('wsUnknownPath');
        const percentageTitle = currentLang === 'de'
            ? `${percentage}% des monatlichen Verbrauchs`
            : `${percentage}% of monthly consumption`;
        
        wsCard.innerHTML = `
            <div class="ws-header">
                <div class="ws-title-block">
                    <span class="ws-name">${escapeHtml(wsSummary.workspace.name)}</span>
                    <span class="ws-path" title="${escapeHtml(pathLabel)}">${escapeHtml(pathLabel)}</span>
                </div>
                <div class="ws-token-block">
                    <span class="ws-token-total">${formatNumber(wsSummary.tokens.total)} ${t('overallTotalTokens')}</span>
                    <span class="ws-token-breakdown">
                        ${t('wsPrompt')}: ${formatNumber(wsSummary.tokens.prompt)} | ${t('wsCompletion')}: ${formatNumber(wsSummary.tokens.completion)}
                    </span>
                    <span class="ws-token-breakdown" style="color: var(--text-primary); font-weight: 500; margin-top: 0.3rem;">
                        ${t('wsPrompts')}: ${formatNumber(wsSummary.tokens.requests)} | ${t('wsCost')}: ${formatCredit(wsSummary.tokens.aic)} AIC (${formatNumber(wsSummary.tokens.aiu)} nano-AIU)
                    </span>
                </div>
            </div>
            <div class="progress-container" title="${percentageTitle}">
                <div class="progress-bar" style="width: ${percentage}%"></div>
            </div>
        `;
        
        workspacesListEl.appendChild(wsCard);
    });
}

function clearDashboard() {
    kpiTotalEl.innerText = '0';
    kpiPromptEl.innerText = '0';
    kpiCompletionEl.innerText = '0';
    kpiRequestsEl.innerText = '0';
    kpiAicEl.innerText = formatCredit(0);
    kpiAiuEl.innerText = '0';
    workspacesListEl.innerHTML = `
        <div class="empty-state">
            <span class="empty-state-icon">🔍</span>
            <p>${t('emptyStateNoData')}</p>
        </div>
    `;
}

// Helper: Format numbers with thousands separator (Locale dependent)
function formatNumber(num) {
    return new Intl.NumberFormat(locale).format(num);
}

// Helper: Format credit values (Locale dependent with 2 decimals)
function formatCredit(num) {
    return new Intl.NumberFormat(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(num);
}

// Helper: Format YYYY-MM to localized month name (Locale dependent)
function formatMonth(monthStr) {
    const parts = monthStr.split('-');
    if (parts.length !== 2) return monthStr;
    
    const year = parts[0];
    const month = parseInt(parts[1], 10) - 1;
    
    const date = new Date(year, month, 1);
    return date.toLocaleDateString(locale, { month: 'long', year: 'numeric' });
}

// Helper: Simple HTML Escaping
function escapeHtml(unsafe) {
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

// About Modal Handlers
window.openAboutModal = function () {
    const modal = document.getElementById('about-modal');
    if (modal) {
        modal.classList.remove('hidden');
    }
};

window.closeAboutModal = function () {
    const modal = document.getElementById('about-modal');
    if (modal) {
        modal.classList.add('hidden');
    }
};

// Close modal when clicking outside of it
window.addEventListener('click', (event) => {
    const modal = document.getElementById('about-modal');
    if (event.target === modal) {
        modal.classList.add('hidden');
    }
});

// Trigger initial scan on startup with timing conditions
window.addEventListener('DOMContentLoaded', () => {
    // Detect language
    const userLang = navigator.language || navigator.userLanguage;
    const isGerman = userLang.startsWith('de');
    currentLang = isGerman ? 'de' : 'en';
    locale = isGerman ? 'de-DE' : 'en-US';

    // Set active language switcher button
    document.querySelectorAll('.lang-btn').forEach(btn => {
        btn.classList.toggle('active', btn.getAttribute('data-lang') === currentLang);
    });

    applyStaticTranslations();

    // Start 5-second timer for the warning banner
    setTimeout(() => {
        isMinTimeElapsed = true;
        checkHideStartupBanner();
    }, 5000);

    // Small delay to ensure Wails runtime is loaded and ready
    setTimeout(async () => {
        try {
            await performInitialScan();
        } catch (err) {
            console.error("Initial scan execution failed:", err);
        } finally {
            isInitialScanFinished = true;
            checkHideStartupBanner();
        }
    }, 100);
});
