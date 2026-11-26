// =====================================================================
// LC Performance Dashboard - frontend logic
//
// Stage 2: data comes from the DUMMY_DATA object below.
// Its shape matches the GraphQL `FunnelSummary` type the Go backend
// will return, so in stage 3 we only replace getFunnel() with a fetch.
// =====================================================================

// Human-readable names for each phase code.
const PHASE_NAMES = {
  SU: "Sign-up",
  APL: "Applied",
  ACC: "Accepted",
  APD: "Approved",
  RE: "Realized",
  FIN: "Finished",
  CO: "Completed",
};

// Fake responses, one per programme. Conversion values are ratios
// (0.4 = 40%) and null means "previous phase was 0" -> shown as "–".
const DUMMY_DATA = {
  OGX: {
    programme: "OGX",
    dataSource: "mock",
    phases: [
      { phase: "SU", count: 420, conversionFromPrevious: null },
      { phase: "APL", count: 168, conversionFromPrevious: 0.4 },
      { phase: "ACC", count: 105, conversionFromPrevious: 0.625 },
      { phase: "APD", count: 63, conversionFromPrevious: 0.6 },
      { phase: "RE", count: 50, conversionFromPrevious: 0.794 },
      { phase: "FIN", count: 42, conversionFromPrevious: 0.84 },
      { phase: "CO", count: 38, conversionFromPrevious: 0.905 },
    ],
    suToApl: 0.4,
    aplToApd: 0.375,
    apdToRe: 0.794,
    suToCo: 0.09,
  },
  ICX: {
    programme: "ICX",
    dataSource: "mock",
    phases: [
      { phase: "SU", count: 0, conversionFromPrevious: null },
      { phase: "APL", count: 85, conversionFromPrevious: null },
      { phase: "ACC", count: 40, conversionFromPrevious: 0.471 },
      { phase: "APD", count: 22, conversionFromPrevious: 0.55 },
      { phase: "RE", count: 18, conversionFromPrevious: 0.818 },
      { phase: "FIN", count: 15, conversionFromPrevious: 0.833 },
      { phase: "CO", count: 14, conversionFromPrevious: 0.933 },
    ],
    // SU is 0 here on purpose, to show how null ("–") is displayed.
    suToApl: null,
    aplToApd: 0.259,
    apdToRe: 0.818,
    suToCo: null,
  },
};

// Currently selected programme ("OGX" or "ICX").
let currentProgramme = "OGX";

// ---------- helpers ----------

// 0.405 -> "40.5%", null -> "–"
function formatPercent(value) {
  if (value === null || value === undefined) return "–";
  return (value * 100).toFixed(1) + "%";
}

// Find the count of one phase, e.g. countOf(data, "SU")
function countOf(data, phaseCode) {
  const row = data.phases.find((p) => p.phase === phaseCode);
  return row ? row.count : 0;
}

// Shortcut for document.getElementById(...).textContent = text
function setText(id, text) {
  document.getElementById(id).textContent = text;
}

// ---------- data ----------

// Stage 2: returns dummy data. Stage 3: will call /graphql.
function getFunnel(programme, startDate, endDate) {
  return DUMMY_DATA[programme];
}

// ---------- rendering ----------

function renderKpis(data) {
  setText("kpi-su", countOf(data, "SU"));
  setText("kpi-apd", countOf(data, "APD"));
  setText("kpi-re", countOf(data, "RE"));
  setText("kpi-suco", formatPercent(data.suToCo));
}

function renderHighlights(data) {
  setText("conv-suapl", formatPercent(data.suToApl));
  setText("conv-aplapd", formatPercent(data.aplToApd));
  setText("conv-apdre", formatPercent(data.apdToRe));
}

function renderTable(data) {
  const tbody = document.getElementById("phase-table");
  tbody.innerHTML = ""; // clear old rows

  data.phases.forEach((p) => {
    const tr = document.createElement("tr");
    tr.innerHTML =
      "<td>" + p.phase + " – " + PHASE_NAMES[p.phase] + "</td>" +
      '<td class="num">' + p.count + "</td>" +
      '<td class="num">' + formatPercent(p.conversionFromPrevious) + "</td>";
    tbody.appendChild(tr);
  });
}

function renderBadge(data) {
  const label = data.dataSource === "expa" ? "Data: EXPA live" : "Data: Mock (dummy)";
  setText("source-badge", label);
}

// Read the filters, get data, and draw everything.
function loadDashboard() {
  const start = document.getElementById("start-date").value;
  const end = document.getElementById("end-date").value;

  const data = getFunnel(currentProgramme, start, end);

  renderBadge(data);
  renderKpis(data);
  renderHighlights(data);
  renderTable(data);
}

// ---------- events ----------

// oGX / iCX toggle buttons
document.querySelectorAll(".toggle button").forEach((button) => {
  button.addEventListener("click", () => {
    // move the "active" class to the clicked button
    document.querySelectorAll(".toggle button").forEach((b) => b.classList.remove("active"));
    button.classList.add("active");

    currentProgramme = button.dataset.programme;
    loadDashboard();
  });
});

document.getElementById("apply-btn").addEventListener("click", loadDashboard);

document.getElementById("export-btn").addEventListener("click", () => {
  alert("CSV export will be connected in a later stage.");
});

// First load when the page opens.
loadDashboard();
