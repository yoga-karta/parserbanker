import { useState, useRef, useEffect } from "react";
const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "/api";

const CATEGORIES = [
  { key: "selisih_kurang", label: "Selisih Kurang", color: "#fb7185", bg: "#2a151b", border: "#4a2430", legend: "Rollback dengan Nominal EJ dan Cash sama" },
  { key: "tidak_ditemukan", label: "Tidak Ditemukan", color: "#a78bfa", bg: "#201a33", border: "#3a2f5c", legend: "Ada di salah satu file tapi tidak ditemukan pasangannya, atau nominal EJ dan Cash tidak sama" },
  { key: "match", label: "Match / Klop", color: "#34d399", bg: "#0f2a22", border: "#1c4a3a", legend: "Data EJ dan Cash sesuai" },
  { key: "data_invalid", label: "Data Rusak", color: "#facc15", bg: "#2a2510", border: "#4a4020", legend: "Nominal EJ atau Cash tidak terbaca (data rusak/kosong)" },
];

function fmtRp(n) {
  if (n === null || n === undefined) return "—";
  return "Rp" + Number(n).toLocaleString("id-ID");
}

// ---------- API helpers (dipakai bareng oleh view Simple & Pro) ----------
async function apiCall(url, options) {
  const res = await fetch(url, options);
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || "Terjadi kesalahan pada server");
  return data;
}
const apiCreateJob = () => apiCall(`${API_BASE}/jobs`, { method: "POST" });
const apiGetJob = (jobId) => apiCall(`${API_BASE}/jobs/${jobId}`);
const apiGetLog = (jobId) => apiCall(`${API_BASE}/jobs/${jobId}/log`);
const apiProcess = (jobId) => apiCall(`${API_BASE}/jobs/${jobId}/process`, { method: "POST" });
const apiStop = (jobId) => apiCall(`${API_BASE}/jobs/${jobId}/stop`, { method: "POST" });
const apiReset = (jobId) => apiCall(`${API_BASE}/jobs/${jobId}/reset`, { method: "POST" });
function apiLoadFile(jobId, role, file, opts) {
  const form = new FormData();
  form.append("file", file);
  form.append("ignore_case", opts.ignoreCase ? "true" : "false");
  form.append("trim_data", opts.trimData ? "true" : "false");
  form.append("skip_header", opts.skipHeader ? "true" : "false");
  form.append("validate_data", opts.validateData ? "true" : "false");
  return apiCall(`${API_BASE}/jobs/${jobId}/load/${role}`, { method: "POST", body: form });
}
function IconAtm(props) {
  return (
    <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <rect x="3" y="4" width="18" height="14" rx="2" />
      <path d="M3 9h18" />
      <path d="M7 13h3M7 15.5h5" />
      <circle cx="17" cy="14" r="1.6" />
    </svg>
  );
}
function IconBank(props) {
  return (
    <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M3 10l9-6 9 6" />
      <path d="M5 10v9M9.5 10v9M14.5 10v9M19 10v9" />
      <path d="M3 19h18" />
    </svg>
  );
}
function IconRefresh(props) {
  return (
    <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M3 12a9 9 0 0 1 15.3-6.4L21 8M21 3v5h-5" />
      <path d="M21 12a9 9 0 0 1-15.3 6.4L3 16M3 21v-5h5" />
    </svg>
  );
}
function IconTrash(props) {
  return (
    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2m3 0-1 14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2L4 6" />
    </svg>
  );
}
function IconChevron({ dir = "left", ...props }) {
  return (
    <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" style={{ transform: dir === "right" ? "rotate(180deg)" : "none" }} {...props}>
      <path d="M15 18l-6-6 6-6" />
    </svg>
  );
}
function IconFolder(props) {
  return (
    <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7Z" />
    </svg>
  );
}
function IconPlay(props) {
  return (
    <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor" {...props}>
      <path d="M8 5v14l11-7z" />
    </svg>
  );
}
function IconStop(props) {
  return (
    <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor" {...props}>
      <rect x="6" y="6" width="12" height="12" rx="1.5" />
    </svg>
  );
}
function IconSearch(props) {
  return (
    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <circle cx="11" cy="11" r="7" />
      <path d="m21 21-4.3-4.3" />
    </svg>
  );
}
function IconExcel(props) {
  return (
    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <rect x="3" y="3" width="18" height="18" rx="2" />
      <path d="m8 8 8 8M16 8l-8 8" />
    </svg>
  );
}
function IconFileText(props) {
  return (
    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
      <path d="M14 2v6h6M8 13h8M8 17h8M8 9h2" />
    </svg>
  );
}
function IconList(props) {
  return (
    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01" />
    </svg>
  );
}
function IconPower(props) {
  return (
    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M12 2v10" />
      <path d="M18.4 6.6a9 9 0 1 1-12.8 0" />
    </svg>
  );
}

function CategoryBadge({ category }) {
  const c = CATEGORIES.find((x) => x.key === category);
  if (!c) return <span>—</span>;
  return (
    <span className="result-badge" style={{ color: c.color, backgroundColor: c.bg, borderColor: c.border }}>
      {c.label}
    </span>
  );
}
function PenyebabBadge({ value }) {
  if (!value) return <span className="dim-text">—</span>;
  return (
    <span className="result-badge" style={{ color: "#f5a524", backgroundColor: "#2a2110", borderColor: "#4a3a1c" }}>
      {value}
    </span>
  );
}
export default function Home() {
  const [viewMode, setViewMode] = useState("simple"); // "simple" | "pro"

  const atmRef = useRef(null);
  const bniRef = useRef(null);
  const [status, setStatus] = useState("idle");
  const [job, setJob] = useState(null);
  const [activeCategory, setActiveCategory] = useState("all");
  const [page, setPage] = useState(1);
  const [results, setResults] = useState({ results: [], total: 0 });
  const [errorMsg, setErrorMsg] = useState("");
  const [atmFileName, setAtmFileName] = useState("");
  const [bniFileName, setBniFileName] = useState("");
  const [historyList, setHistoryList] = useState([]);

  // ---- state khusus tampilan Pro ----
  const proEjRef = useRef(null);
  const proRcRef = useRef(null);
  const [proJobId, setProJobId] = useState(null);
  const [proEjInfo, setProEjInfo] = useState({ name: "", count: null });
  const [proRcInfo, setProRcInfo] = useState({ name: "", count: null });
  const [proOptions, setProOptions] = useState({ ignoreCase: false, trimData: true, skipHeader: false, validateData: true });
  const [proLoadingEj, setProLoadingEj] = useState(false);
  const [proLoadingRc, setProLoadingRc] = useState(false);
  const [proErrorMsg, setProErrorMsg] = useState("");
  const [proLog, setProLog] = useState([]);
  const [showLogModal, setShowLogModal] = useState(false);
  const [previewModal, setPreviewModal] = useState(null); // { title, lines }

  useEffect(() => {
    fetchHistory();
  }, []);

  async function fetchHistory() {
    try {
      const res = await fetch(`${API_BASE}/history`);
      if (res.ok) {
        const data = await res.json();
        setHistoryList((data.history || []).reverse());
      }
    } catch (e) {
      console.error("Gagal memuat riwayat:", e);
    }
  }

  async function handleClearHistory() {
    if (!confirm("Apakah Anda yakin ingin menghapus seluruh riwayat komparasi?")) return;
    try {
      const res = await fetch(`${API_BASE}/history`, { method: "DELETE" });
      if (res.ok) {
        setHistoryList([]);
        handleReset();
      }
    } catch (e) {
      alert("Gagal menghapus riwayat log");
    }
  }
  // ---------- Simple view: 1 tombol, orchestrate create->load->load->process di belakang layar ----------
  async function handleSubmit(e) {
    e.preventDefault();
    setErrorMsg("");
    const atmFile = atmRef.current.files[0];
    const bniFile = bniRef.current.files[0];
    if (!atmFile || !bniFile) {
      setErrorMsg("Kedua file data (EJ & Cash) wajib diunggah untuk komparasi.");
      return;
    }

    setStatus("uploading");
    try {
      const defaultOpts = { ignoreCase: false, trimData: true, skipHeader: false, validateData: true };
      const created = await apiCreateJob();
      const jobId = created.job_id;
      await apiLoadFile(jobId, "ej", atmFile, defaultOpts);
      await apiLoadFile(jobId, "rc", bniFile, defaultOpts);
      setStatus("processing");
      await apiProcess(jobId);
      pollJob(jobId);
    } catch (err) {
      setErrorMsg(err.message);
      setStatus("error");
    }
  }

  function pollJob(jobId) {
    const interval = setInterval(async () => {
      try {
        const data = await apiGetJob(jobId);
        setJob(data);
        if (data.status === "done") {
          clearInterval(interval);
          setStatus("done");
          // Tampilkan SEMUA kategori dulu begitu proses selesai - operator butuh
          // lihat gambaran lengkap sebelum nge-filter ke kategori tertentu.
          loadResults(jobId, "all", 1);
          fetchHistory();
        } else if (data.status === "error") {
          clearInterval(interval);
          setStatus("error");
          setErrorMsg(data.error || "Terjadi kesalahan saat kalkulasi data");
        }
      } catch (err) {
        clearInterval(interval);
        setStatus("error");
        setErrorMsg("Gagal menghubungi server rekonsiliasi");
      }
    }, 1500);
  }

  async function handleSelectHistory(item) {
    if (item.status !== "done") return;
    setErrorMsg("");
    setJob({ id: item.job_id, status: item.status, summary: item.summary || {} });
    setStatus("done");
    await loadResults(item.job_id, "all", 1);
  }

  async function loadResults(jobId, category, pageNum) {
    try {
      const res = await fetch(
        `${API_BASE}/jobs/${jobId}/results?category=${category}&page=${pageNum}&page_size=50`
      );
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || "Data hasil kalkulasi sesi ini tidak ditemukan / sudah kadaluarsa di RAM server.");
      }
      setResults({
        results: Array.isArray(data.results) ? data.results : [],
        total: typeof data.total === "number" ? data.total : 0,
      });
      setActiveCategory(category);
      setPage(pageNum);
      setErrorMsg("");
    } catch (err) {
      setErrorMsg(err.message);
      setResults({ results: [], total: 0 });
    }
  }

  function handleReset() {
    setStatus("idle");
    setJob(null);
    setResults({ results: [], total: 0 });
    setAtmFileName("");
    setBniFileName("");
    if (atmRef.current) atmRef.current.value = "";
    if (bniRef.current) bniRef.current.value = "";
    // reset state Pro juga biar konsisten pas pindah balik ke Simple
    setProJobId(null);
    setProEjInfo({ name: "", count: null });
    setProRcInfo({ name: "", count: null });
    setProErrorMsg("");
    setProLog([]);
    if (proEjRef.current) proEjRef.current.value = "";
    if (proRcRef.current) proRcRef.current.value = "";
  }
  // ---------- Pro view: 2 langkah eksplisit (Load -> Process), opsi fungsional, Stop, Log ----------
  async function handleProLoad(role) {
    const fileRef = role === "ej" ? proEjRef : proRcRef;
    const file = fileRef.current.files[0];
    if (!file) {
      setProErrorMsg(`Pilih file ${role === "ej" ? "EJ" : "RC"} dulu sebelum Load Data.`);
      return;
    }
    setProErrorMsg("");
    role === "ej" ? setProLoadingEj(true) : setProLoadingRc(true);
    try {
      let jobId = proJobId;
      if (!jobId) {
        const created = await apiCreateJob();
        jobId = created.job_id;
        setProJobId(jobId);
      }
      const data = await apiLoadFile(jobId, role, file, proOptions);
      if (role === "ej") {
        setProEjInfo({ name: file.name, count: data.total_record });
      } else {
        setProRcInfo({ name: file.name, count: data.total_record });
      }
    } catch (err) {
      setProErrorMsg(err.message);
    } finally {
      role === "ej" ? setProLoadingEj(false) : setProLoadingRc(false);
    }
  }

  async function handleProProcess() {
    if (!proJobId || proEjInfo.count == null || proRcInfo.count == null) {
      setProErrorMsg("Load kedua file (Data 1 & Data 2) dulu sebelum Process/Recon.");
      return;
    }
    setProErrorMsg("");
    setStatus("processing");
    try {
      await apiProcess(proJobId);
      pollJob(proJobId);
    } catch (err) {
      setProErrorMsg(err.message);
      setStatus("error");
    }
  }

  async function handleProStop() {
    if (!proJobId) return;
    try {
      await apiStop(proJobId);
      setStatus("idle");
      setProErrorMsg("Proses dihentikan (STOP).");
    } catch (err) {
      setProErrorMsg(err.message);
    }
  }

  async function handleProReset() {
    if (proJobId) {
      try {
        await apiReset(proJobId);
      } catch (e) {
        // job mungkin sudah tidak ada, aman diabaikan
      }
    }
    handleReset();
  }

  async function handleProViewLog() {
    if (!proJobId) {
      setProLog([]);
      setShowLogModal(true);
      return;
    }
    try {
      const data = await apiGetLog(proJobId);
      setProLog(data.log || []);
    } catch (err) {
      setProLog([]);
    }
    setShowLogModal(true);
  }

  // Preview murni client-side (baca beberapa baris awal file langsung dari browser,
  // tidak perlu round-trip ke server / gak nyimpen data sensitif di mana-mana)
  function handleProPreview(role) {
    const fileRef = role === "ej" ? proEjRef : proRcRef;
    const file = fileRef.current.files[0];
    if (!file) {
      setProErrorMsg(`Belum ada file ${role === "ej" ? "EJ" : "RC"} yang dipilih untuk di-preview.`);
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const lines = String(reader.result).split(/\r?\n/).slice(0, 30);
      setPreviewModal({ title: `Preview ${role === "ej" ? "FILE TXT 1 - EJ" : "FILE TXT 2 - RC"} (${file.name})`, lines });
    };
    reader.readAsText(file.slice(0, 200 * 1024)); // baca max 200KB pertama, cukup buat preview
  }

  // category: "all" (default) export semua baris, atau salah satu key CATEGORIES
  // buat export cuma kategori yang lagi difilter.
  function exportUrl(format, category) {
    const jobId = viewMode === "pro" ? proJobId : job?.id;
    const cat = category && category !== "all" ? `&category=${category}` : "";
    return `${API_BASE}/jobs/${jobId}/export?format=${format}${cat}`;
  }

  const s = job?.summary || {};
  const allCount = (s.match || 0) + (s.selisih_kurang || 0) + (s.tidak_ditemukan || 0) + (s.data_invalid || 0);
  return (
    <div className={viewMode === "pro" ? "layout pro-mode" : "layout"}>
      <header className="header">
        <div className="header-content">
          <div className="brand">
            <div className="logo-badge">BNI</div>
            <div>
              <h1><span className="live-dot" aria-hidden="true"></span>Reconciliation Pilot Portal</h1>
              <p>Automated Transaction Matching Engine &bull; ATM vs Core BNI</p>
            </div>
          </div>
          <div className="user-nav">
            <div className="view-toggle">
              <button className={viewMode === "simple" ? "active" : ""} onClick={() => setViewMode("simple")}>Simple</button>
              <button className={viewMode === "pro" ? "active" : ""} onClick={() => setViewMode("pro")}>Pro</button>
            </div>
            {status === "done" && (
              <button onClick={handleReset} className="btn-ghost">
                <IconRefresh /> Komparasi Baru
              </button>
            )}
            <button onClick={handleReset} className="btn-reset">
              <IconRefresh /> Reset
            </button>
          </div>
        </div>
      </header>

      {viewMode === "simple" ? (
      <main className="container">
        {errorMsg && (
          <div className="alert-error">
            <strong>Error:</strong> {errorMsg}
          </div>
        )}

        {historyList.length > 0 && (
          <div className="card history-card" style={{ marginBottom: 24 }}>
            <div className="card-header history-card-header">
              <div>
                <h2>Riwayat Komparasi</h2>
                <p>Klik salah satu untuk memuat kembali hasilnya.</p>
              </div>
              <button onClick={handleClearHistory} className="btn-clear-history">
                <IconTrash /> Hapus Semua
              </button>
            </div>
            <div className="history-list">
              {historyList.map((h, i) => {
                const perluCek = (h.counts?.selisih_kurang || 0) + (h.counts?.tidak_ditemukan || 0);
                return (
                  <div
                    key={h.job_id}
                    onClick={() => handleSelectHistory(h)}
                    className={`history-item ${job?.id === h.job_id ? "active-history" : ""}`}
                  >
                    <div className="history-info">
                      <strong>Pencocokan #{historyList.length - i}</strong>
                      <span className="history-date">
                        {new Date(h.timestamp).toLocaleString("id-ID")}
                      </span>
                    </div>
                    {h.status === "done" ? (
                      <div className="history-stats">
                        <span className="badge-match">Match: {h.counts?.match || 0}</span>
                        <span className="badge-diff">Perlu Cek: {perluCek}</span>
                      </div>
                    ) : (
                      <span className="badge-error">Error / Gagal</span>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {status !== "done" && (
          <div className="card upload-card">
            <div className="card-header">
              <h2>Unggah Berkas Transaksi</h2>
              <p>Format yang didukung: Electronic Journal &amp; Cash/Reconciliation Data (.txt)</p>
            </div>

            <form onSubmit={handleSubmit}>
              <div className="dropzones">
                <div className="dropzone-box" onClick={() => atmRef.current.click()}>
                  <div className="dz-icon dz-icon-blue"><IconAtm /></div>
                  <div className="label-title">Electronic Journal / Transaction Data</div>
                  <div className="label-file">
                    {atmFileName || "Klik untuk memilih berkas EJ (.txt)"}
                  </div>
                  <input
                    type="file"
                    ref={atmRef}
                    accept=".txt,.log"
                    style={{ display: "none" }}
                    onChange={(e) => setAtmFileName(e.target.files[0]?.name || "")}
                  />
                </div>

                <div className="dropzone-box" onClick={() => bniRef.current.click()}>
                  <div className="dz-icon dz-icon-orange"><IconBank /></div>
                  <div className="label-title">Cash / Reconciliation Data</div>
                  <div className="label-file">
                    {bniFileName || "Klik untuk memilih berkas Cash/Reconciliation (.txt)"}
                  </div>
                  <input
                    type="file"
                    ref={bniRef}
                    accept=".txt"
                    style={{ display: "none" }}
                    onChange={(e) => setBniFileName(e.target.files[0]?.name || "")}
                  />
                </div>
              </div>

              <div className="action-area">
                <button
                  type="submit"
                  className="btn-primary"
                  disabled={status === "uploading" || status === "processing"}
                >
                  {status === "uploading"
                    ? "Mengunggah Berkas..."
                    : status === "processing"
                    ? "Mencocokkan Rec Num..."
                    : "Jalankan Rekonsiliasi Otomatis"}
                </button>
              </div>

              {(status === "uploading" || status === "processing") && (
                <div className="status-banner">
                  <div className="spinner"></div>
                  <span>
                    {status === "uploading"
                      ? "Sedang mengunggah file secara streaming ke memori server..."
                      : "Engine DuckDB sedang mencocokkan Rec Num EJ vs Cash..."}
                  </span>
                </div>
              )}
            </form>
          </div>
        )}

        {job && job.status === "done" && (
          <div className="results-section">
            <div className="totals-row">
              <div className="total-card">
                <span className="nominal-label">Total Transaksi (EJ)</span>
                <span className="kpi-value">{(s.total_ej || 0).toLocaleString("id-ID")}</span>
              </div>
              <div className="total-card">
                <span className="nominal-label">Total Data Cash (CRM)</span>
                <span className="kpi-value">{(s.total_cash || 0).toLocaleString("id-ID")}</span>
              </div>
            </div>

            <div className="kpi-grid">
              {CATEGORIES.map((c) => {
                const count = s[c.key] ?? 0;
                const isActive = activeCategory === c.key;
                return (
                  <div
                    key={c.key}
                    onClick={() => loadResults(job.id, c.key, 1)}
                    className={`kpi-card ${isActive ? "active" : ""}`}
                    style={{
                      borderColor: isActive ? c.color : "#232f4a",
                      backgroundColor: isActive ? c.bg : "#111a2e",
                    }}
                  >
                    <span className="kpi-label" style={{ color: c.color }}>
                      {c.label}
                    </span>
                    <span className="kpi-value">{count.toLocaleString("id-ID")}</span>
                    <span className="kpi-sub">baris transaksi</span>
                  </div>
                );
              })}
            </div>

            <div className="nominal-grid">
              <div className="nominal-card">
                <span className="nominal-label">Total Nominal EJ</span>
                <span className="nominal-value">{fmtRp(s.nominal_ej)}</span>
              </div>
              <div className="nominal-card">
                <span className="nominal-label">Total Nominal Cash</span>
                <span className="nominal-value">{fmtRp(s.nominal_cash)}</span>
              </div>
              <div className="nominal-card accent-red">
                <span className="nominal-label">Selisih Nominal (Kurang)</span>
                <span className="nominal-value">{fmtRp(s.selisih_nominal_kurang)}</span>
              </div>
            </div>

            <div className="card table-card">
              <div className="table-header">
                <div>
                  <h3>Detail Transaksi: {activeCategory === "all" ? "Semua Kategori" : CATEGORIES.find((c) => c.key === activeCategory)?.label}</h3>
                  <p>
                    Menampilkan total <strong>{(results.total || 0).toLocaleString("id-ID")}</strong> baris data (dicocokkan via Rec Num)
                  </p>
                </div>
                <div className="pagination-top">
                  <button
                    onClick={() => loadResults(job.id, "all", 1)}
                    className={`btn-ghost ${activeCategory === "all" ? "btn-ghost-active" : ""}`}
                  >
                    Semua Kategori
                  </button>
                  <a href={exportUrl("xlsx", activeCategory)} className="btn-ghost" target="_blank" rel="noreferrer">
                    Export Excel {activeCategory !== "all" ? "(filter ini)" : "(semua)"}
                  </a>
                  <a href={exportUrl("txt", activeCategory)} className="btn-ghost" target="_blank" rel="noreferrer">
                    Export TXT {activeCategory !== "all" ? "(filter ini)" : "(semua)"}
                  </a>
                  <button
                    className="btn-page"
                    disabled={page <= 1}
                    onClick={() => loadResults(job.id, activeCategory, page - 1)}
                  >
                    <IconChevron dir="left" />
                  </button>
                  <span className="page-indicator">
                    Hal. {page} / {Math.max(1, Math.ceil((results.total || 0) / 50))}
                  </span>
                  <button
                    className="btn-page"
                    disabled={page * 50 >= (results.total || 0)}
                    onClick={() => loadResults(job.id, activeCategory, page + 1)}
                  >
                    <IconChevron dir="right" />
                  </button>
                </div>
              </div>

              <div className="table-responsive">
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>No</th>
                      <th>Tanggal</th>
                      <th>Terminal</th>
                      <th>Nomor Kartu</th>
                      <th>Nomor Rekening</th>
                      <th>Rec Num</th>
                      <th>Nominal EJ</th>
                      <th>Nominal Cash</th>
                      <th>EJ Status</th>
                      <th>Cash Status</th>
                      <th>Result</th>
                      <th>Kemungkinan Penyebab</th>
                      <th>Keterangan</th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.results && results.results.length > 0 ? (
                      results.results.map((r, i) => (
                        <tr key={`${r.rec_num}-${i}`}>
                          <td>{(page - 1) * 50 + i + 1}</td>
                          <td>{r.tanggal || "—"}</td>
                          <td>{r.terminal || "—"}</td>
                          <td className="font-mono">{r.no_rekening || "—"}</td>
                          <td className="font-mono">{r.no_transaksi || "—"}</td>
                          <td><span className="key-pill">{r.rec_num}</span></td>
                          <td className="font-mono">{fmtRp(r.nominal_ej)}</td>
                          <td className="font-mono">{fmtRp(r.nominal_cash)}</td>
                          <td>{r.ej_status || "—"}</td>
                          <td>{r.cash_status || "—"}</td>
                          <td><CategoryBadge category={r.category} /></td>
                          <td><PenyebabBadge value={r.kemungkinan_penyebab} /></td>
                          <td className="keterangan-cell">{r.keterangan || "-"}</td>
                        </tr>
                      ))
                    ) : (
                      <tr>
                        <td colSpan="13" className="empty-state">
                          Tidak ada transaksi pada kategori ini.
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}
      </main>
      ) : (
      <main className="pro-container">
        <div className="pro-panel">
          <h2 className="pro-panel-title">INPUT DATA</h2>
          <div className="pro-input-grid">
            <div className="pro-file-box">
              <div className="pro-file-label pro-label-blue">FILE TXT 1 - EJ / TRANSACTION DATA</div>
              <div className="pro-file-row">
                <input type="text" readOnly value={proEjInfo.name} placeholder="Belum ada file dipilih" />
                <button type="button" className="pro-btn-browse" onClick={() => proEjRef.current.click()}>
                  <IconFolder /> Browse
                </button>
                <input type="file" ref={proEjRef} accept=".txt,.log" style={{ display: "none" }}
                  onChange={(e) => setProEjInfo({ name: e.target.files[0]?.name || "", count: null })} />
              </div>
              <div className="pro-file-meta">
                <span>Format&nbsp;&nbsp;: Raw Log ATM (Electronic Journal, multi-baris)</span>
                <span>Total Data : {proEjInfo.count != null ? `${proEjInfo.count.toLocaleString("id-ID")} record` : "-"}</span>
              </div>
            </div>

            <div className="pro-file-box">
              <div className="pro-file-label pro-label-green">FILE TXT 2 - CASH / RECONCILIATION DATA</div>
              <div className="pro-file-row">
                <input type="text" readOnly value={proRcInfo.name} placeholder="Belum ada file dipilih" />
                <button type="button" className="pro-btn-browse" onClick={() => proRcRef.current.click()}>
                  <IconFolder /> Browse
                </button>
                <input type="file" ref={proRcRef} accept=".txt" style={{ display: "none" }}
                  onChange={(e) => setProRcInfo({ name: e.target.files[0]?.name || "", count: null })} />
              </div>
              <div className="pro-file-meta">
                <span>Format&nbsp;&nbsp;: Settlement RC (semicolon-delimited)</span>
                <span>Total Data : {proRcInfo.count != null ? `${proRcInfo.count.toLocaleString("id-ID")} record` : "-"}</span>
              </div>
            </div>
          </div>

          <div className="pro-actions-row">
            <button type="button" className="pro-btn" onClick={() => handleProLoad("ej")} disabled={proLoadingEj}>
              <IconFolder /> {proLoadingEj ? "Memuat..." : "LOAD DATA 1"}
            </button>
            <button type="button" className="pro-btn" onClick={() => handleProLoad("rc")} disabled={proLoadingRc}>
              <IconFolder /> {proLoadingRc ? "Memuat..." : "LOAD DATA 2"}
            </button>
            <button type="button" className="pro-btn pro-btn-process"
              onClick={handleProProcess}
              disabled={proEjInfo.count == null || proRcInfo.count == null || status === "processing"}>
              <IconPlay /> PROCESS / RECON
            </button>
            <button type="button" className="pro-btn pro-btn-stop" onClick={handleProStop} disabled={status !== "processing"}>
              <IconStop /> STOP
            </button>
            <button type="button" className="pro-btn pro-btn-reset" onClick={handleProReset}>
              <IconRefresh /> RESET
            </button>
          </div>

          <div className="pro-status-row">
            <span className="pro-status-label">STATUS PROSES</span>
            <span className={`pro-status-text ${status === "done" ? "pro-status-done" : status === "error" ? "pro-status-err" : ""}`}>
              {status === "processing" ? "Sedang memproses..." : status === "done" ? "Proses selesai!" : status === "error" ? "Terjadi error" : "Menunggu file di-load"}
            </span>
            <div className="pro-progress-track">
              <div className="pro-progress-fill" style={{ width: status === "done" ? "100%" : status === "processing" ? "60%" : "0%" }} />
            </div>
          </div>

          {proErrorMsg && <div className="pro-alert">{proErrorMsg}</div>}
          {errorMsg && <div className="pro-alert">{errorMsg}</div>}
        </div>

        {job && job.status === "done" && (
          <>
            <div className="pro-panel">
              <div className="pro-panel-header">
                <h2 className="pro-panel-title">HASIL REKONSILIASI</h2>
                <div className="pro-filter-group">
                  <button
                    className={`pro-filter-btn ${activeCategory === "all" ? "pro-filter-active" : ""}`}
                    onClick={() => loadResults(job.id, "all", 1)}
                  >
                    Semua ({allCount.toLocaleString("id-ID")})
                  </button>
                  {CATEGORIES.map((c) => (
                    <button
                      key={c.key}
                      className={`pro-filter-btn ${activeCategory === c.key ? "pro-filter-active" : ""}`}
                      style={activeCategory === c.key ? { borderColor: c.color, color: c.color } : undefined}
                      onClick={() => loadResults(job.id, c.key, 1)}
                    >
                      {c.label} ({(s[c.key] ?? 0).toLocaleString("id-ID")})
                    </button>
                  ))}
                  <a
                    className="pro-filter-btn pro-filter-export"
                    href={exportUrl("xlsx", activeCategory)}
                    target="_blank" rel="noreferrer"
                    title={activeCategory === "all" ? "Export semua kategori" : `Export cuma kategori ${CATEGORIES.find((c) => c.key === activeCategory)?.label}`}
                  >
                    <IconExcel /> Export {activeCategory === "all" ? "Semua" : "Filter Ini"}
                  </a>
                </div>
              </div>
              <div className="pro-table-wrap">
                <table className="pro-table">
                  <thead>
                    <tr>
                      <th>No</th><th>Tanggal</th><th>Terminal</th><th>Nomor Kartu</th><th>Nomor Rekening</th>
                      <th className="pro-th-recnum">Rec Num</th><th>Nominal EJ</th><th>Nominal Cash</th>
                      <th>EJ Status</th><th>Cash Status</th><th>Result</th><th>Keterangan</th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.results && results.results.length > 0 ? (
                      results.results.map((r, i) => (
                        <tr key={`${r.rec_num}-${i}`} className={`pro-row-${r.category}`}>
                          <td>{(page - 1) * 50 + i + 1}</td>
                          <td>{r.tanggal || "—"}</td>
                          <td>{r.terminal || "—"}</td>
                          <td>{r.no_rekening || "—"}</td>
                          <td>{r.no_transaksi || "—"}</td>
                          <td className="pro-th-recnum">{r.rec_num}</td>
                          <td>{fmtRp(r.nominal_ej)}</td>
                          <td>{fmtRp(r.nominal_cash)}</td>
                          <td className={r.ej_status === "SUCCESS" ? "pro-text-green" : r.ej_status === "ROLLBACK" ? "pro-text-red" : ""}>{r.ej_status}</td>
                          <td>{r.cash_status}</td>
                          <td><span className={`pro-badge pro-badge-${r.category}`}>{CATEGORIES.find((c) => c.key === r.category)?.label.toUpperCase() || r.category}</span></td>
                          <td>{r.keterangan}</td>
                        </tr>
                      ))
                    ) : (
                      <tr><td colSpan="12" className="pro-empty">Tidak ada data.</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
              <div className="pro-pagination">
                <button className="pro-btn" disabled={page <= 1} onClick={() => loadResults(job.id, activeCategory, page - 1)}><IconChevron dir="left" /></button>
                <span>Hal. {page} / {Math.max(1, Math.ceil((results.total || 0) / 50))} &bull; {(results.total || 0).toLocaleString("id-ID")} baris</span>
                <button className="pro-btn" disabled={page * 50 >= (results.total || 0)} onClick={() => loadResults(job.id, activeCategory, page + 1)}><IconChevron dir="right" /></button>
              </div>
            </div>

            <div className="pro-summary-grid">
              <div className="pro-panel">
                <h2 className="pro-panel-title">RINGKASAN HASIL</h2>
                <div className="pro-stat-row">
                  <div className="pro-stat"><IconFileText /><div><span>Total Transaksi (EJ)</span><strong>{(s.total_ej || 0).toLocaleString("id-ID")}</strong></div></div>
                  <div className="pro-stat"><IconFileText /><div><span>Total Data Cash (CRM)</span><strong>{(s.total_cash || 0).toLocaleString("id-ID")}</strong></div></div>
                  <div className="pro-stat pro-stat-green"><IconSearch /><div><span>Klop / Match</span><strong>{(s.match || 0).toLocaleString("id-ID")}</strong></div></div>
                  <div className="pro-stat pro-stat-red"><IconSearch /><div><span>Selisih Kurang</span><strong>{(s.selisih_kurang || 0).toLocaleString("id-ID")}</strong></div></div>
                  <div className="pro-stat pro-stat-purple"><IconSearch /><div><span>Tidak Ditemukan</span><strong>{(s.tidak_ditemukan || 0).toLocaleString("id-ID")}</strong></div></div>
                  <div className="pro-stat pro-stat-yellow"><IconSearch /><div><span>Data Rusak</span><strong>{(s.data_invalid || 0).toLocaleString("id-ID")}</strong></div></div>
                </div>
                <div className="pro-nominal-row">
                  <div><span>Total Nominal EJ</span><strong>{fmtRp(s.nominal_ej)}</strong></div>
                  <div><span>Total Nominal Cash</span><strong>{fmtRp(s.nominal_cash)}</strong></div>
                  <div><span>Selisih Nominal (Kurang)</span><strong className="pro-text-red">{fmtRp(s.selisih_nominal_kurang)}</strong></div>
                </div>
              </div>

              <div className="pro-panel">
                <h2 className="pro-panel-title">DETAIL RESULT</h2>
                <div className="pro-legend">
                  {CATEGORIES.map((c) => (
                    <div key={c.key} className="pro-legend-row">
                      <span className="pro-legend-dot" style={{ background: c.color }} />
                      <strong>{c.label.toUpperCase()}</strong>
                      <span>: {c.legend}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="pro-bottom-bar">
              <button className="pro-btn" onClick={() => handleProPreview("ej")}><IconSearch /> PREVIEW DATA 1</button>
              <button className="pro-btn" onClick={() => handleProPreview("rc")}><IconSearch /> PREVIEW DATA 2</button>
              <a className="pro-btn pro-btn-excel" href={exportUrl("xlsx", "all")} target="_blank" rel="noreferrer" title="Export semua kategori, ditandai per baris"><IconExcel /> EXPORT EXCEL (SEMUA)</a>
              <a className="pro-btn" href={exportUrl("txt", "all")} target="_blank" rel="noreferrer" title="Export semua kategori, ditandai per baris"><IconFileText /> EXPORT TXT (SEMUA)</a>
              <button className="pro-btn" onClick={handleProViewLog}><IconList /> VIEW LOG</button>
              <button className="pro-btn pro-btn-exit" onClick={handleReset}><IconPower /> EXIT</button>
            </div>
          </>
        )}

        <div className="pro-statusbar">
          <span>Aplikasi Pengolahan 2 Data TXT Berbeda Format &bull; Web Edition</span>
        </div>
      </main>
      )}

      {showLogModal && (
        <div className="modal-overlay" onClick={() => setShowLogModal(false)}>
          <div className="modal-box" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header"><h3>View Log</h3><button onClick={() => setShowLogModal(false)}>×</button></div>
            <div className="modal-body">
              {proLog.length === 0 ? (
                <p className="dim-text">Belum ada aktivitas.</p>
              ) : (
                proLog.map((l, i) => (
                  <div key={i} className="log-entry">
                    <span className="log-time">{new Date(l.timestamp).toLocaleTimeString("id-ID")}</span>
                    <span>{l.message}</span>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      )}

      {previewModal && (
        <div className="modal-overlay" onClick={() => setPreviewModal(null)}>
          <div className="modal-box modal-wide" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header"><h3>{previewModal.title}</h3><button onClick={() => setPreviewModal(null)}>×</button></div>
            <div className="modal-body">
              <pre className="preview-pre">{previewModal.lines.join("\n")}</pre>
            </div>
          </div>
        </div>
      )}
      <style jsx>{`
        .layout {
          font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
          min-height: 100vh;
          background: #0a0e1a;
          color: #eef2fa;
        }
        .layout.pro-mode { background: #eef1f6; color: #1e293b; }
        .header {
          background: #0d1425;
          border-bottom: 1px solid #1c2740;
          padding: 16px 0;
        }
        .header-content {
          max-width: 1220px; margin: 0 auto; padding: 0 24px;
          display: flex; justify-content: space-between; align-items: center;
        }
        .brand { display: flex; align-items: center; gap: 16px; }
        .logo-badge {
          background: linear-gradient(135deg, #f2711c, #d85c10);
          color: white; font-weight: 700; padding: 8px 13px; border-radius: 8px;
          font-size: 15px; letter-spacing: 0.5px; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
          box-shadow: 0 6px 16px -6px rgba(242, 113, 28, 0.5);
        }
        .brand h1 {
          margin: 0; font-size: 17px; font-weight: 700; color: #f2f5fb;
          font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; display: flex; align-items: center; gap: 8px;
        }
        .live-dot {
          width: 7px; height: 7px; border-radius: 50%; background: #34d399; display: inline-block;
          box-shadow: 0 0 0 rgba(52, 211, 153, 0.5);
        }
        @media (prefers-reduced-motion: no-preference) {
          .live-dot { animation: pulse 2s ease-in-out infinite; }
          @keyframes pulse {
            0%, 100% { box-shadow: 0 0 0 0 rgba(52, 211, 153, 0.5); }
            50% { box-shadow: 0 0 0 5px rgba(52, 211, 153, 0); }
          }
        }
        .brand p { margin: 3px 0 0; font-size: 12px; color: #6b7791; }
        .user-nav { display: flex; align-items: center; gap: 10px; }
        .view-toggle {
          display: flex; background: #0d1425; border: 1px solid #263252; border-radius: 8px; padding: 3px; gap: 2px;
        }
        .view-toggle button {
          background: transparent; border: none; color: #7c88a4; padding: 6px 14px; border-radius: 6px;
          font-size: 12.5px; font-weight: 600; cursor: pointer; transition: all 0.15s;
        }
        .view-toggle button.active { background: #f2711c; color: white; }
        .btn-ghost {
          display: inline-flex; align-items: center; gap: 6px;
          background: #16213a; color: #dbe2f0; border: 1px solid #263252;
          padding: 7px 13px; border-radius: 7px; cursor: pointer; font-weight: 500; font-size: 12.5px;
          transition: background 0.15s; text-decoration: none;
        }
        .btn-ghost:hover { background: #1c2942; }
        .btn-ghost-active { background: #1c2942; border-color: #37507e; }
        .btn-reset {
          display: inline-flex; align-items: center; gap: 6px;
          background: rgba(251, 113, 133, 0.12); color: #fca5b3; border: 1px solid rgba(251, 113, 133, 0.3);
          padding: 7px 13px; border-radius: 7px; cursor: pointer; font-weight: 600; font-size: 12.5px;
          transition: background 0.15s;
        }
        .btn-reset:hover { background: rgba(251, 113, 133, 0.2); }
        .container { max-width: 1220px; margin: 32px auto; padding: 0 24px 60px; }
        .alert-error {
          background: rgba(251, 113, 133, 0.1); border: 1px solid rgba(251, 113, 133, 0.3); color: #fca5b3;
          padding: 12px 16px; border-radius: 10px; margin-bottom: 24px; font-size: 14px;
        }
        .card {
          background: #111a2e; border-radius: 14px; border: 1px solid #1c2740;
          box-shadow: 0 1px 0 rgba(255, 255, 255, 0.02) inset; overflow: hidden;
        }
        .card-header { padding: 22px 24px 4px; }
        .card-header h2 { margin: 0 0 5px; font-size: 18px; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; color: #f2f5fb; }
        .card-header p { margin: 0; color: #7c88a4; font-size: 13.5px; }
        .history-card-header { display: flex; justify-content: space-between; align-items: flex-start; padding-bottom: 16px; }
        .btn-clear-history {
          display: inline-flex; align-items: center; gap: 6px;
          background: rgba(251, 113, 133, 0.1); color: #fca5b3; border: 1px solid rgba(251, 113, 133, 0.3);
          padding: 6px 13px; border-radius: 7px; cursor: pointer; font-weight: 600; font-size: 11.5px; transition: background 0.15s;
        }
        .btn-clear-history:hover { background: rgba(251, 113, 133, 0.18); }
        .history-list { display: flex; gap: 12px; overflow-x: auto; padding: 0 24px 22px; }
        .history-item {
          min-width: 200px; padding: 13px 16px; border: 1px solid #1c2740; border-radius: 10px;
          cursor: pointer; background: #0d1425; transition: all 0.15s;
        }
        .history-item:hover { border-color: #37507e; background: #101b31; }
        .active-history { border-color: #f2711c; background: #1c1712; }
        .history-info { display: flex; flex-direction: column; margin-bottom: 9px; }
        .history-info strong { font-size: 13px; color: #eef2fa; }
        .history-date { font-size: 11px; color: #6b7791; margin-top: 2px; }
        .history-stats { display: flex; gap: 8px; font-size: 11px; font-weight: 600; }
        .badge-diff { background: rgba(251, 113, 133, 0.12); color: #fb7185; padding: 3px 7px; border-radius: 5px; }
        .badge-match { background: rgba(52, 211, 153, 0.12); color: #34d399; padding: 3px 7px; border-radius: 5px; }
        .badge-error { background: rgba(251, 113, 133, 0.12); color: #fb7185; font-size: 11px; font-weight: 600; padding: 3px 7px; border-radius: 5px; }
        .dropzones { display: grid; grid-template-columns: 1fr 1fr; gap: 18px; padding: 22px 24px; }
        .dropzone-box {
          border: 1.5px dashed #263252; border-radius: 12px; padding: 30px 18px; text-align: center;
          cursor: pointer; transition: all 0.15s; background: #0d1425;
        }
        .dropzone-box:hover { border-color: #f2711c; background: #14100c; }
        .dz-icon {
          width: 46px; height: 46px; margin: 0 auto 12px; border-radius: 10px;
          display: flex; align-items: center; justify-content: center;
        }
        .dz-icon-blue { background: rgba(91, 155, 255, 0.12); color: #5b9bff; }
        .dz-icon-orange { background: rgba(242, 113, 28, 0.12); color: #f2711c; }
        .label-title { font-weight: 600; margin-bottom: 6px; color: #dbe2f0; font-size: 14px; }
        .label-file { font-size: 12.5px; color: #6b7791; word-break: break-all; }
        .action-area { padding: 4px 24px 24px; text-align: right; }
        .btn-primary {
          background: linear-gradient(135deg, #f2711c, #d85c10); color: white; border: none;
          padding: 12px 26px; border-radius: 9px; font-weight: 600; font-size: 14.5px; cursor: pointer;
          transition: filter 0.15s, transform 0.1s;
        }
        .btn-primary:hover:not(:disabled) { filter: brightness(1.08); }
        .btn-primary:active:not(:disabled) { transform: translateY(1px); }
        .btn-primary:disabled { background: #263252; color: #6b7791; cursor: not-allowed; filter: none; }
        .status-banner {
          display: flex; align-items: center; gap: 12px; background: #0d1425; border-top: 1px solid #1c2740;
          padding: 14px 24px; color: #9fb0d1; font-size: 13.5px; font-weight: 500;
        }
        .spinner {
          width: 17px; height: 17px; border: 2.5px solid #263252; border-top: 2.5px solid #f2711c;
          border-radius: 50%; animation: spin 0.8s linear infinite; flex-shrink: 0;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
        .kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; margin-bottom: 14px; }
        .kpi-card {
          padding: 18px 20px; border-radius: 12px; border: 1.5px solid #232f4a;
          cursor: pointer; transition: all 0.15s; display: flex; flex-direction: column;
        }
        .kpi-card:hover { transform: translateY(-2px); }
        .kpi-label { font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.4px; margin-bottom: 8px; }
        .kpi-value { font-size: 26px; font-weight: 700; color: #f2f5fb; line-height: 1; }
        .kpi-sub { font-size: 11.5px; color: #6b7791; margin-top: 6px; }
        .totals-row { display: grid; grid-template-columns: repeat(2, 1fr); gap: 14px; margin-bottom: 14px; }
        .total-card {
          padding: 16px 20px; border-radius: 12px; background: #111a2e; border: 1.5px solid #232f4a;
          display: flex; flex-direction: column; gap: 6px;
        }
        .result-badge {
          display: inline-block; font-size: 11px; font-weight: 700; padding: 3px 9px;
          border-radius: 20px; border: 1px solid; white-space: nowrap;
        }
        .dim-text { color: #4a5674; }
        .nominal-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; margin-bottom: 22px; }
        .nominal-card {
          padding: 14px 18px; border-radius: 10px; background: #111a2e; border: 1px solid #1c2740;
          display: flex; flex-direction: column; gap: 4px;
        }
        .nominal-card.accent-red { border-color: #4a2430; }
        .nominal-card.accent-blue { border-color: #233a5c; }
        .nominal-label { font-size: 11px; color: #7c88a4; text-transform: uppercase; letter-spacing: 0.3px; }
        .nominal-value { font-size: 16px; font-weight: 700; color: #eef2fa; font-family: ui-monospace, monospace; }
        .nominal-card.accent-red .nominal-value { color: #fb7185; }
        .nominal-card.accent-blue .nominal-value { color: #5b9bff; }
        .table-header {
          padding: 20px 24px; border-bottom: 1px solid #1c2740; display: flex;
          justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 10px;
        }
        .table-header h3 { margin: 0 0 4px; font-size: 16.5px; color: #f2f5fb; }
        .table-header p { margin: 0; font-size: 12.5px; color: #6b7791; }
        .pagination-top { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
        .btn-page {
          display: flex; align-items: center; justify-content: center;
          background: #0d1425; border: 1px solid #263252; color: #dbe2f0; width: 32px; height: 32px;
          border-radius: 7px; cursor: pointer; transition: background 0.15s;
        }
        .btn-page:hover:not(:disabled) { background: #16213a; }
        .btn-page:disabled { opacity: 0.35; cursor: not-allowed; }
        .page-indicator { font-size: 12.5px; font-weight: 600; color: #9fb0d1; padding: 0 4px; }
        .table-responsive { overflow-x: auto; }
        .data-table { width: 100%; border-collapse: collapse; text-align: left; font-size: 12.5px; white-space: nowrap; }
        .data-table th {
          background: #0d1425; padding: 10px 16px; font-weight: 600; color: #7c88a4;
          border-bottom: 1px solid #1c2740; font-size: 11px; text-transform: uppercase; letter-spacing: 0.3px;
        }
        .data-table td { padding: 11px 16px; border-bottom: 1px solid #161f36; vertical-align: top; }
        .data-table tr:hover td { background: #0d1425; }
        .font-mono { font-family: ui-monospace, monospace; }
        .keterangan-cell { color: #9fb0d1; white-space: normal; min-width: 180px; }
        .key-pill {
          font-family: ui-monospace, monospace; font-size: 12px; font-weight: 600; color: #f2711c;
          background: rgba(242, 113, 28, 0.1); padding: 3px 8px; border-radius: 5px;
        }
        .empty-state { text-align: center; padding: 44px !important; color: #4a5674; font-style: italic; white-space: normal; }

        /* ================= PRO VIEW (tema terang) ================= */
        .pro-container { max-width: 1400px; margin: 0 auto; padding: 20px 24px 60px; }
        .pro-panel {
          background: #ffffff; border: 1px solid #d7dce5; border-radius: 8px;
          padding: 18px 20px; margin-bottom: 16px; box-shadow: 0 1px 2px rgba(0,0,0,0.03);
        }
        .pro-panel-title { margin: 0 0 14px; font-size: 14px; font-weight: 700; color: #1e3a8a; letter-spacing: 0.3px; }
        .pro-panel-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; flex-wrap: wrap; }
        .pro-panel-header .pro-panel-title { margin: 0; }
        .pro-filter-group { display: flex; gap: 6px; flex-wrap: wrap; justify-content: flex-end; }
        .pro-filter-btn {
          display: inline-flex; align-items: center; gap: 5px; background: #ffffff; border: 1px solid #cbd5e1;
          color: #475569; padding: 5px 10px; border-radius: 5px; font-size: 11.5px; font-weight: 700;
          cursor: pointer; text-decoration: none; white-space: nowrap;
        }
        .pro-filter-btn:hover { background: #f1f5f9; }
        .pro-filter-active { background: #eff6ff; border-color: #93c5fd; color: #1d4ed8; }
        .pro-filter-export { background: #f0fdf4; border-color: #86efac; color: #15803d; }
        .pro-filter-export:hover { background: #dcfce7; }
        .pro-input-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
        .pro-file-box { border: 1px solid #e2e8f0; border-radius: 6px; padding: 12px 14px; }
        .pro-file-label { font-size: 12px; font-weight: 700; margin-bottom: 8px; color: #334155; }
        .pro-label-blue { color: #1d4ed8; }
        .pro-label-green { color: #15803d; }
        .pro-file-row { display: flex; gap: 8px; margin-bottom: 8px; }
        .pro-file-row input[type="text"] {
          flex: 1; padding: 8px 10px; border: 1px solid #cbd5e1; border-radius: 5px; font-size: 12.5px;
          color: #334155; background: #f8fafc;
        }
        .pro-btn-browse {
          display: inline-flex; align-items: center; gap: 5px; background: #f1f5f9; border: 1px solid #cbd5e1;
          color: #334155; padding: 7px 12px; border-radius: 5px; font-size: 12px; font-weight: 600; cursor: pointer;
          white-space: nowrap;
        }
        .pro-btn-browse:hover { background: #e2e8f0; }
        .pro-file-meta { display: flex; flex-direction: column; gap: 3px; font-size: 11.5px; color: #64748b; }
        .pro-actions-row { display: flex; gap: 10px; flex-wrap: wrap; margin-bottom: 14px; }
        .pro-btn {
          display: inline-flex; align-items: center; gap: 6px; background: #ffffff; border: 1px solid #cbd5e1;
          color: #334155; padding: 9px 16px; border-radius: 6px; font-size: 12.5px; font-weight: 700; cursor: pointer;
          transition: background 0.15s; text-decoration: none;
        }
        .pro-btn:hover:not(:disabled) { background: #f1f5f9; }
        .pro-btn:disabled { opacity: 0.45; cursor: not-allowed; }
        .pro-btn-process { background: #16a34a; color: white; border-color: #16a34a; }
        .pro-btn-process:hover:not(:disabled) { background: #15803d; }
        .pro-btn-stop { color: #64748b; }
        .pro-btn-reset { color: #ea580c; border-color: #fed7aa; }
        .pro-btn-excel { background: #16a34a; color: white; border-color: #16a34a; }
        .pro-btn-exit { color: #dc2626; border-color: #fecaca; }
        .pro-status-row { display: flex; align-items: center; gap: 14px; }
        .pro-status-label { font-size: 12px; font-weight: 700; color: #334155; white-space: nowrap; }
        .pro-status-text { font-size: 12.5px; font-weight: 600; color: #64748b; white-space: nowrap; }
        .pro-status-done { color: #16a34a; }
        .pro-status-err { color: #dc2626; }
        .pro-progress-track { flex: 1; height: 8px; background: #e2e8f0; border-radius: 6px; overflow: hidden; }
        .pro-progress-fill { height: 100%; background: linear-gradient(90deg, #16a34a, #22c55e); transition: width 0.3s; }
        .pro-alert { margin-top: 12px; background: #fef2f2; border: 1px solid #fecaca; color: #b91c1c; padding: 10px 14px; border-radius: 6px; font-size: 13px; }
        .pro-table-wrap { overflow-x: auto; border: 1px solid #e2e8f0; border-radius: 6px; }
        .pro-table { width: 100%; border-collapse: collapse; font-size: 12px; white-space: nowrap; }
        .pro-table th {
          background: #f8fafc; color: #334155; text-align: left; padding: 9px 12px; font-weight: 700;
          border-bottom: 2px solid #e2e8f0; position: sticky; top: 0;
        }
        .pro-table td { padding: 8px 12px; border-bottom: 1px solid #eef1f6; color: #334155; }
        .pro-th-recnum { border-left: 2px solid #ef4444; border-right: 2px solid #ef4444; }
        .pro-row-match td { background: #ecfdf5; }
        .pro-row-selisih_kurang td { background: #fef2f2; }
        .pro-row-tidak_ditemukan td { background: #f5f3ff; }
        .pro-row-data_invalid td { background: #fefce8; }
        .pro-text-green { color: #16a34a; font-weight: 600; }
        .pro-text-red { color: #dc2626; font-weight: 600; }
        .pro-text-blue { color: #2563eb; }
        .pro-badge { font-size: 10.5px; font-weight: 700; padding: 3px 8px; border-radius: 4px; white-space: nowrap; }
        .pro-badge-match { background: #16a34a; color: white; }
        .pro-badge-selisih_kurang { background: #dc2626; color: white; }
        .pro-badge-tidak_ditemukan { background: #7c3aed; color: white; }
        .pro-badge-data_invalid { background: #ca8a04; color: white; }
        .pro-empty { text-align: center; padding: 30px !important; color: #94a3b8; font-style: italic; white-space: normal; }
        .pro-pagination { display: flex; align-items: center; gap: 10px; margin-top: 12px; font-size: 12.5px; color: #64748b; }
        .pro-summary-grid { display: grid; grid-template-columns: 1.4fr 1fr; gap: 16px; margin-bottom: 16px; }
        .pro-stat-row { display: flex; gap: 12px; flex-wrap: wrap; margin-bottom: 14px; }
        .pro-stat {
          display: flex; align-items: center; gap: 10px; border: 1px solid #e2e8f0; border-radius: 6px;
          padding: 10px 14px; flex: 1; min-width: 130px; color: #94a3b8;
        }
        .pro-stat div { display: flex; flex-direction: column; }
        .pro-stat span { font-size: 10.5px; color: #64748b; }
        .pro-stat strong { font-size: 17px; color: #1e293b; }
        .pro-stat-green { color: #16a34a; }
        .pro-stat-red { color: #dc2626; }
        .pro-stat-purple { color: #7c3aed; }
        .pro-stat-yellow { color: #ca8a04; }
        .pro-nominal-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; padding-top: 12px; border-top: 1px solid #eef1f6; }
        .pro-nominal-row > div { display: flex; flex-direction: column; gap: 3px; }
        .pro-nominal-row span { font-size: 10.5px; color: #64748b; }
        .pro-nominal-row strong { font-size: 14px; color: #1e293b; font-family: ui-monospace, monospace; }
        .pro-legend { display: flex; flex-direction: column; gap: 10px; }
        .pro-legend-row { display: flex; align-items: center; gap: 8px; font-size: 12.5px; color: #334155; }
        .pro-legend-dot { width: 12px; height: 12px; border-radius: 3px; flex-shrink: 0; }
        .pro-bottom-bar { display: flex; gap: 10px; flex-wrap: wrap; padding: 14px 4px; }
        .pro-statusbar {
          display: flex; justify-content: space-between; font-size: 11px; color: #94a3b8;
          padding: 10px 4px; border-top: 1px solid #d7dce5; margin-top: 8px;
        }

        .modal-overlay {
          position: fixed; inset: 0; background: rgba(10, 14, 26, 0.6); display: flex;
          align-items: center; justify-content: center; z-index: 50; padding: 20px;
        }
        .modal-box { background: #111a2e; border: 1px solid #232f4a; border-radius: 12px; width: 100%; max-width: 480px; max-height: 80vh; display: flex; flex-direction: column; }
        .modal-wide { max-width: 700px; }
        .modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid #1c2740; }
        .modal-header h3 { margin: 0; font-size: 15px; color: #f2f5fb; }
        .modal-header button { background: none; border: none; color: #7c88a4; font-size: 20px; cursor: pointer; line-height: 1; }
        .modal-body { padding: 16px 20px; overflow-y: auto; }
        .log-entry { display: flex; gap: 12px; font-size: 12.5px; color: #dbe2f0; padding: 6px 0; border-bottom: 1px solid #1c2740; }
        .log-time { color: #6b7791; font-family: ui-monospace, monospace; white-space: nowrap; }
        .preview-pre { font-size: 11.5px; color: #dbe2f0; white-space: pre-wrap; word-break: break-all; font-family: ui-monospace, monospace; }
      `}</style>
    </div>
  );
}
