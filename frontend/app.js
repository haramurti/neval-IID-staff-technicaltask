// =====================================================================
// LC Performance Dashboard - frontend logic
//
// Data still comes from DUMMY_DATA below. Its shape matches the GraphQL
// `FunnelSummary` type the Go backend will return, so in stage 3 we
// only replace the inside of getFunnel() with a fetch to /graphql.
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

// The Chart.js chart object. We keep it so we can update it later
// instead of creating a new chart every time.
let chart = null;

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

// True when every phase count is 0 -> nothing to show.
function isEmpty(data) {
  return data.phases.every((p) => p.count === 0);
}

// ---------- status (loading / empty / error) ----------

// type: "loading" | "empty" | "error" | null (null = hide the box)
function showStatus(type, message) {
  const box = document.getElementById("status");
  const results = document.getElementById("results");

  if (type === null) {
    box.hidden = true;
    results.hidden = false;
    return;
  }

  box.className = "status " + type;
  box.textContent = message;
  box.hidden = false;
  results.hidden = true; // hide numbers so old data is never mistaken for new
}

// Returns an error message, or null if the dates are fine.
function validateDates(start, end) {
  if (!start || !end) return "Please fill in both start and end date.";
  // "YYYY-MM-DD" strings can be compared directly as text.
  if (start > end) return "Start date must be before or equal to end date.";
  return null;
}

// ---------- data ----------

// Returns a Promise, like fetch() will in stage 3.
// For now it waits a bit to imitate a network call, then returns dummy data.
//
// Test helpers (dummy only, removed in stage 3):
//   - end date before 2026-01-01  -> empty result
//   - add ?fail to the URL        -> simulated error
function getFunnel(programme, startDate, endDate) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      if (location.search.includes("fail")) {
        reject(new Error("Could not reach the server (simulated)."));
        return;
      }

      const data = DUMMY_DATA[programme];

      if (endDate < "2026-01-01") {
        // copy the data with every count set to 0
        resolve({
          ...data,
          phases: data.phases.map((p) => ({ ...p, count: 0, conversionFromPrevious: null })),
          suToApl: null, aplToApd: null, apdToRe: null, suToCo: null,
        });
        return;
      }

      resolve(data);
    }, 500);
  });
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

function renderChart(data) {
  const labels = data.phases.map((p) => p.phase);
  const counts = data.phases.map((p) => p.count);

  // Chart already exists: just swap the numbers and redraw.
  if (chart) {
    chart.data.labels = labels;
    chart.data.datasets[0].data = counts;
    chart.update();
    return;
  }

  // First time: create the chart.
  chart = new Chart(document.getElementById("phase-chart"), {
    type: "bar",
    data: {
      labels: labels,
      datasets: [{
        label: "Count",
        data: counts,
        backgroundColor: "#E3343A",
      }],
    },
    options: {
      maintainAspectRatio: false, // fill the .chart-box height
      plugins: { legend: { display: false } },
      scales: { y: { beginAtZero: true, ticks: { precision: 0 } } },
    },
  });
}

function renderBadge(data) {
  const label = data.dataSource === "expa" ? "Data: EXPA live" : "Data: Mock (dummy)";
  setText("source-badge", label);
}

// Read the filters, get data, and draw everything.
async function loadDashboard() {
  const start = document.getElementById("start-date").value;
  const end = document.getElementById("end-date").value;

  const dateError = validateDates(start, end);
  if (dateError) {
    showStatus("error", dateError);
    return;
  }

  showStatus("loading", "Loading data…");

  try {
    const data = await getFunnel(currentProgramme, start, end);

    renderBadge(data);

    if (isEmpty(data)) {
      showStatus("empty", "No data for this period.");
      return;
    }

    // Show the results first, so Chart.js can measure its box size.
    showStatus(null);
    renderKpis(data);
    renderHighlights(data);
    renderTable(data);
    renderChart(data);
  } catch (err) {
    showStatus("error", "Failed to load data: " + err.message);
  }
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
  alert("CSV export will be connected once the backend is ready.");
});

// First load when the page opens.
loadDashboard();
