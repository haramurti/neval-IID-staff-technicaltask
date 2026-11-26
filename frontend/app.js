// =====================================================================
// LC Performance Dashboard - frontend logic
//
// Data comes from our Go server's GraphQL endpoint (/graphql).
// The browser never talks to EXPA directly: the server does that,
// so the EXPA token never reaches this file.
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

// The GraphQL query. It asks for exactly the fields the page uses.
// $programme, $startDate, $endDate are variables sent next to the query.
const FUNNEL_QUERY = `
  query Funnel($programme: Programme!, $startDate: String!, $endDate: String!) {
    funnel(programme: $programme, startDate: $startDate, endDate: $endDate) {
      programme
      startDate
      endDate
      dataSource
      suToApl
      aplToApd
      apdToRe
      suToCo
      phases {
        phase
        count
        conversionFromPrevious
      }
    }
  }
`;

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

// Calls POST /graphql and returns the `funnel` object.
//
// A GraphQL request is always a POST with JSON: { query, variables }.
// A GraphQL response is { data, errors }. Errors (e.g. bad dates) usually
// come back with HTTP 200, so we must check `errors` ourselves.
async function getFunnel(programme, startDate, endDate) {
  const response = await fetch("/graphql", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      query: FUNNEL_QUERY,
      variables: { programme: programme, startDate: startDate, endDate: endDate },
    }),
  });

  if (!response.ok) {
    throw new Error("Server returned HTTP " + response.status);
  }

  const result = await response.json();

  if (result.errors && result.errors.length > 0) {
    throw new Error(result.errors[0].message);
  }

  return result.data.funnel;
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
