import { useState, useRef, useEffect } from "react";
const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "/api";

const CATEGORIES = [
  { key: "selisih_kurang", label: "Selisih Kurang", color: "#dc2626", bg: "#fef2f2", border: "#fecaca", legend: "Rollback dengan Nominal EJ dan Cash sama" },
  { key: "tidak_ditemukan", label: "Tidak Ditemukan", color: "#7c3aed", bg: "#f5f3ff", border: "#ddd6fe", legend: "Ada di salah satu file tapi tidak ditemukan pasangannya, atau nominal EJ dan Cash tidak sama" },
  { key: "match", label: "Match / Klop", color: "#059669", bg: "#ecfdf5", border: "#a7f3d0", legend: "Data EJ dan Cash sesuai" },
  { key: "data_invalid", label: "Data Rusak", color: "#d97706", bg: "#fffbeb", border: "#fde68a", legend: "Nominal EJ atau Cash tidak terbaca (data rusak/kosong)" },
];

function fmtRp(n) {
  if (n === null || n === undefined) return "—";
  return "Rp" + Number(n).toLocaleString("id-ID");
}

// ---------- API helpers ----------
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
    <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <rect x="3" y="4" width="18" height="14" rx="2" />
      <path d="M3 9h18" />
      <path d="M7 13h3M7 15.5h5" />
      <circle cx="17" cy="14" r="1.6" />
    </svg>
  );
}
function IconBank(props) {
  return (
    <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" {...props}>
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
    <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01" />
    </svg>
  );
}

function CategoryBadge({ category }) {
  const c = CATEGORIES.find((x) => x.key === category);
  if (!c) return <span>—</span>;
  return (
    <span className="badge" style={{ color: c.color, backgroundColor: c.bg, borderColor: c.border }}>
      {c.label}
    </span>
  );
}
function PenyebabBadge({ value }) {
  if (!value) return <span className="dim-text">—</span>;
  return (
    <span className="badge" style={{ color: "#b45309", backgroundColor: "#fffbeb", borderColor: "#fde68a" }}>
      {value}
    </span>
  );
}

function FileSlot({ label, hint, icon, iconClass, info, loading, inputRef, accept, onChange, onBrowse, onLoad, onPreview }) {
  const loaded = info.count != null;
  return (
    <div className={`file-slot ${loaded ? "file-slot-loaded" : ""}`}>
      <div className="file-slot-head">
        <div className={`file-slot-icon ${iconClass}`}>{icon}</div>
        <div>
          <div className="file-slot-label">{label}</div>
          <div className="file-slot-hint">{hint}</div>
        </div>
      </div>
      <div className="file-slot-row">
        <input type="text" readOnly value={info.name} placeholder="Belum ada berkas dipilih" onClick={onBrowse} />
        <button type="button" className="btn-chip" onClick={onBrowse}><IconFolder /> Pilih</button>
        <input type="file" ref={inputRef} accept={accept} style={{ display: "none" }} onChange={onChange} />
      </div>
      <div className="file-slot-footer">
        <span className="file-slot-count">{loaded ? `${info.count.toLocaleString("id-ID")} baris terbaca` : "—"}</span>
        <div className="file-slot-actions">
          <button type="button" className="btn-link" onClick={onPreview} disabled={!info.name}>Preview</button>
          <button type="button" className="btn-chip" onClick={onLoad} disabled={loading || !info.name}>{loading ? "Memuat…" : loaded ? "Muat Ulang" : "Muat Berkas"}</button>
        </div>
      </div>
    </div>
  );
}

function SidePanel({ panel, onClose, historyList, activeJobId, onSelectHistory, onClearHistory }) {
  if (!panel) return null;
  const isHistory = panel.type === "history";
  const row = panel.data;
  return (
    <div className="panel-overlay" onClick={onClose}>
      <aside className="side-panel" onClick={(e) => e.stopPropagation()}>
        <div className="side-panel-head">
          <h3>{isHistory ? "Riwayat Rekonsiliasi" : "Detail Transaksi"}</h3>
          <button className="panel-close" onClick={onClose}>×</button>
        </div>

        {isHistory ? (
          <div className="side-panel-body">
            {historyList.length === 0 ? (
              <p className="dim-text">Belum ada riwayat.</p>
            ) : (
              <>
                <div className="history-list-v">
                  {historyList.map((h, i) => {
                    const perluCek = (h.counts?.selisih_kurang || 0) + (h.counts?.tidak_ditemukan || 0);
                    return (
                      <div
                        key={h.job_id}
                        onClick={() => onSelectHistory(h)}
                        className={`history-row ${activeJobId === h.job_id ? "history-row-active" : ""}`}
                      >
                        <div className="history-row-top">
                          <strong>Pencocokan #{historyList.length - i}</strong>
                          <span>{new Date(h.timestamp).toLocaleString("id-ID")}</span>
                        </div>
                        {h.status === "done" ? (
                          <div className="history-row-stats">
                            <span className="pill pill-green">Match {h.counts?.match || 0}</span>
                            <span className="pill pill-red">Perlu Cek {perluCek}</span>
                          </div>
                        ) : (
                          <span className="pill pill-red">Error / Gagal</span>
                        )}
                      </div>
                    );
                  })}
                </div>
                <button className="btn-link-danger" onClick={onClearHistory}><IconTrash /> Hapus Semua Riwayat</button>
              </>
            )}
          </div>
        ) : (
          <div className="side-panel-body">
            <div className="detail-amounts">
              <div>
                <span>Nominal EJ</span>
                <strong>{fmtRp(row.nominal_ej)}</strong>
              </div>
              <div>
                <span>Nominal Cash</span>
                <strong>{fmtRp(row.nominal_cash)}</strong>
              </div>
            </div>
            <CategoryBadge category={row.category} />
            <div className="detail-rows">
              <div className="detail-row"><span>Rec Num</span><strong>{row.rec_num}</strong></div>
              <div className="detail-row"><span>Tanggal</span><strong>{row.tanggal || "—"}</strong></div>
              <div className="detail-row"><span>Terminal</span><strong>{row.terminal || "—"}</strong></div>
              <div className="detail-row"><span>Nomor Kartu</span><strong>{row.no_rekening || "—"}</strong></div>
              <div className="detail-row"><span>Nomor Rekening</span><strong>{row.no_transaksi || "—"}</strong></div>
              <div className="detail-row"><span>EJ Status</span><strong>{row.ej_status || "—"}</strong></div>
              <div className="detail-row"><span>Cash Status</span><strong>{row.cash_status || "—"}</strong></div>
            </div>
            <div className="detail-note">
              <span>Keterangan</span>
              <p>{row.keterangan || "-"}</p>
            </div>
            {row.kemungkinan_penyebab && (
              <div className="detail-note">
                <span>Kemungkinan Penyebab</span>
                <PenyebabBadge value={row.kemungkinan_penyebab} />
              </div>
            )}
          </div>
        )}
      </aside>
    </div>
  );
}

export default function Home() {
  const [status, setStatus] = useState("idle");
  const [job, setJob] = useState(null);
  const [activeCategory, setActiveCategory] = useState("all");
  const [page, setPage] = useState(1);
  const [results, setResults] = useState({ results: [], total: 0 });
  const [errorMsg, setErrorMsg] = useState("");
  const [historyList, setHistoryList] = useState([]);
  const [panel, setPanel] = useState(null); // null | {type:"history"} | {type:"row", data}

  const ejRef = useRef(null);
  const rcRef = useRef(null);
  const [jobId, setJobId] = useState(null);
  const [ejInfo, setEjInfo] = useState({ name: "", count: null });
  const [rcInfo, setRcInfo] = useState({ name: "", count: null });
  const [loadOptions] = useState({ ignoreCase: false, trimData: true, skipHeader: false, validateData: true });
  const [loadingEj, setLoadingEj] = useState(false);
  const [loadingRc, setLoadingRc] = useState(false);
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

  function pollJob(id) {
    const interval = setInterval(async () => {
      try {
        const data = await apiGetJob(id);
        setJob(data);
        if (data.status === "done") {
          clearInterval(interval);
          setStatus("done");
          loadResults(id, "all", 1);
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
    setJobId(item.job_id);
    setJob({ id: item.job_id, status: item.status, summary: item.summary || {} });
    setStatus("done");
    setPanel(null);
    await loadResults(item.job_id, "all", 1);
  }

  async function loadResults(id, category, pageNum) {
    try {
      const res = await fetch(
        `${API_BASE}/jobs/${id}/results?category=${category}&page=${pageNum}&page_size=50`
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
    setPanel(null);
    setJobId(null);
    setEjInfo({ name: "", count: null });
    setRcInfo({ name: "", count: null });
    setErrorMsg("");
    setProLog([]);
    if (ejRef.current) ejRef.current.value = "";
    if (rcRef.current) rcRef.current.value = "";
  }

  async function handleFullReset() {
    if (jobId) {
      try {
        await apiReset(jobId);
      } catch (e) {
        // job mungkin sudah tidak ada di server, aman diabaikan
      }
    }
    handleReset();
  }

  async function handleLoad(role) {
    const fileRef = role === "ej" ? ejRef : rcRef;
    const file = fileRef.current.files[0];
    if (!file) {
      setErrorMsg(`Pilih file ${role === "ej" ? "EJ" : "RC"} dulu sebelum dimuat.`);
      return;
    }
    setErrorMsg("");
    role === "ej" ? setLoadingEj(true) : setLoadingRc(true);
    try {
      let id = jobId;
      if (!id) {
        const created = await apiCreateJob();
        id = created.job_id;
        setJobId(id);
      }
      const data = await apiLoadFile(id, role, file, loadOptions);
      if (role === "ej") {
        setEjInfo({ name: file.name, count: data.total_record });
      } else {
        setRcInfo({ name: file.name, count: data.total_record });
      }
    } catch (err) {
      setErrorMsg(err.message);
    } finally {
      role === "ej" ? setLoadingEj(false) : setLoadingRc(false);
    }
  }

  async function handleProcess() {
    if (!jobId || ejInfo.count == null || rcInfo.count == null) {
      setErrorMsg("Muat kedua berkas (EJ & RC) dulu sebelum menjalankan rekonsiliasi.");
      return;
    }
    setErrorMsg("");
    setStatus("processing");
    try {
      await apiProcess(jobId);
      pollJob(jobId);
    } catch (err) {
      setErrorMsg(err.message);
      setStatus("error");
    }
  }

  async function handleStop() {
    if (!jobId) return;
    try {
      await apiStop(jobId);
      setStatus("idle");
      setErrorMsg("Proses dihentikan (STOP).");
    } catch (err) {
      setErrorMsg(err.message);
    }
  }

  async function handleViewLog() {
    if (!jobId) {
      setProLog([]);
      setShowLogModal(true);
      return;
    }
    try {
      const data = await apiGetLog(jobId);
      setProLog(data.log || []);
    } catch (err) {
      setProLog([]);
    }
    setShowLogModal(true);
  }

  // Preview murni client-side (baca beberapa baris awal file langsung dari browser).
  function handlePreview(role) {
    const fileRef = role === "ej" ? ejRef : rcRef;
    const file = fileRef.current.files[0];
    if (!file) {
      setErrorMsg(`Belum ada file ${role === "ej" ? "EJ" : "RC"} yang dipilih untuk di-preview.`);
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const lines = String(reader.result).split(/\r?\n/).slice(0, 30);
      setPreviewModal({ title: `Preview ${role === "ej" ? "EJ / Transaction Data" : "Cash / Reconciliation Data"} (${file.name})`, lines });
    };
    reader.readAsText(file.slice(0, 200 * 1024));
  }

  function exportUrl(format, category) {
    const cat = category && category !== "all" ? `&category=${category}` : "";
    return `${API_BASE}/jobs/${jobId}/export?format=${format}${cat}`;
  }

  const s = job?.summary || {};
  const allCount = (s.match || 0) + (s.selisih_kurang || 0) + (s.tidak_ditemukan || 0) + (s.data_invalid || 0);
  const isDone = job && job.status === "done";
  const ready = ejInfo.count != null && rcInfo.count != null;

  return (
    <div className="app">
      <header className="topbar">
        <div className="topbar-inner">
          <div className="brand">
            <div className="logo-badge">BNI</div>
            <div>
              <h1>Reconciliation Portal</h1>
              <p>Automated Transaction Matching &bull; ATM vs Core Banking</p>
            </div>
          </div>
          <div className="topbar-actions">
            <button className="btn-outline" onClick={() => setPanel({ type: "history" })}>
              <IconList /> Riwayat{historyList.length > 0 && <span className="count-pill">{historyList.length}</span>}
            </button>
            {isDone && (
              <button className="btn-primary-outline" onClick={handleFullReset}>
                <IconRefresh /> Rekonsiliasi Baru
              </button>
            )}
          </div>
        </div>
      </header>

      <main className="main">
        {errorMsg && <div className="alert">{errorMsg}</div>}

        {!isDone && (
          <section className="card upload-card">
            <div className="card-head">
              <h2>Unggah Berkas Transaksi</h2>
              <p>Muat kedua berkas untuk mulai rekonsiliasi otomatis</p>
            </div>
            <div className="upload-grid">
              <FileSlot
                label="Electronic Journal"
                hint="Raw log ATM, multi-baris (.txt)"
                icon={<IconAtm />}
                iconClass="icon-blue"
                info={ejInfo}
                loading={loadingEj}
                inputRef={ejRef}
                accept=".txt,.log"
                onChange={(e) => setEjInfo({ name: e.target.files[0]?.name || "", count: null })}
                onBrowse={() => ejRef.current.click()}
                onLoad={() => handleLoad("ej")}
                onPreview={() => handlePreview("ej")}
              />
              <FileSlot
                label="Cash / Reconciliation"
                hint="Settlement RC, semicolon-delimited (.txt)"
                icon={<IconBank />}
                iconClass="icon-orange"
                info={rcInfo}
                loading={loadingRc}
                inputRef={rcRef}
                accept=".txt"
                onChange={(e) => setRcInfo({ name: e.target.files[0]?.name || "", count: null })}
                onBrowse={() => rcRef.current.click()}
                onLoad={() => handleLoad("rc")}
                onPreview={() => handlePreview("rc")}
              />
            </div>
            <div className="upload-footer">
              <div className="status-line">
                <span className={`status-dot status-${status}`} />
                <span>
                  {status === "processing" ? "Mencocokkan Rec Num EJ vs Cash…" : status === "error" ? "Terjadi error, cek pesan di atas" : ready ? "Siap dijalankan" : "Menunggu kedua berkas dimuat"}
                </span>
              </div>
              <div className="upload-actions">
                {status === "processing" ? (
                  <button className="btn-outline-danger" onClick={handleStop}><IconStop /> Stop</button>
                ) : (
                  <button className="btn-primary" disabled={!ready || status === "processing"} onClick={handleProcess}>
                    <IconPlay /> Jalankan Rekonsiliasi
                  </button>
                )}
              </div>
            </div>
          </section>
        )}

        {isDone && (
          <>
            <section className="kpi-row">
              <div className="kpi-card kpi-card-plain">
                <span className="kpi-label">Total EJ</span>
                <strong className="kpi-value">{(s.total_ej || 0).toLocaleString("id-ID")}</strong>
              </div>
              <div className="kpi-card kpi-card-plain">
                <span className="kpi-label">Total Cash</span>
                <strong className="kpi-value">{(s.total_cash || 0).toLocaleString("id-ID")}</strong>
              </div>
              {CATEGORIES.map((c) => {
                const active = activeCategory === c.key;
                return (
                  <button
                    key={c.key}
                    className={`kpi-card ${active ? "kpi-card-active" : ""}`}
                    style={active ? { borderColor: c.color } : undefined}
                    onClick={() => loadResults(jobId, c.key, 1)}
                  >
                    <span className="kpi-label" style={{ color: c.color }}>{c.label}</span>
                    <strong className="kpi-value">{(s[c.key] ?? 0).toLocaleString("id-ID")}</strong>
                  </button>
                );
              })}
            </section>

            <div className="nominal-strip">
              <div><span>Total Nominal EJ</span><strong>{fmtRp(s.nominal_ej)}</strong></div>
              <div><span>Total Nominal Cash</span><strong>{fmtRp(s.nominal_cash)}</strong></div>
              <div><span>Selisih Nominal (Kurang)</span><strong className="text-red">{fmtRp(s.selisih_nominal_kurang)}</strong></div>
            </div>

            <section className="card results-card">
              <div className="results-header">
                <div className="filter-pills">
                  <button className={`pill-btn ${activeCategory === "all" ? "pill-btn-active" : ""}`} onClick={() => loadResults(jobId, "all", 1)}>
                    Semua ({allCount.toLocaleString("id-ID")})
                  </button>
                  {CATEGORIES.map((c) => (
                    <button
                      key={c.key}
                      className={`pill-btn ${activeCategory === c.key ? "pill-btn-active" : ""}`}
                      style={activeCategory === c.key ? { borderColor: c.color, color: c.color } : undefined}
                      onClick={() => loadResults(jobId, c.key, 1)}
                    >
                      {c.label} ({(s[c.key] ?? 0).toLocaleString("id-ID")})
                    </button>
                  ))}
                </div>
                <div className="results-actions">
                  <a className="btn-outline" href={exportUrl("xlsx", activeCategory)} target="_blank" rel="noreferrer">
                    <IconExcel /> Excel
                  </a>
                  <a className="btn-outline" href={exportUrl("txt", activeCategory)} target="_blank" rel="noreferrer">
                    <IconFileText /> TXT
                  </a>
                </div>
              </div>

              <div className="table-wrap">
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>No</th><th>Tanggal</th><th>Terminal</th><th>Nomor Kartu</th><th>Nomor Rekening</th>
                      <th>Rec Num</th><th>Nominal EJ</th><th>Nominal Cash</th><th>EJ Status</th><th>Cash Status</th><th>Result</th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.results && results.results.length > 0 ? (
                      results.results.map((r, i) => (
                        <tr key={`${r.rec_num}-${i}`} onClick={() => setPanel({ type: "row", data: r })}>
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
                        </tr>
                      ))
                    ) : (
                      <tr><td colSpan="11" className="empty-state">Tidak ada transaksi pada kategori ini.</td></tr>
                    )}
                  </tbody>
                </table>
              </div>

              <div className="pagination">
                <button className="btn-page" disabled={page <= 1} onClick={() => loadResults(jobId, activeCategory, page - 1)}><IconChevron dir="left" /></button>
                <span>Hal. {page} / {Math.max(1, Math.ceil((results.total || 0) / 50))} &bull; {(results.total || 0).toLocaleString("id-ID")} baris</span>
                <button className="btn-page" disabled={page * 50 >= (results.total || 0)} onClick={() => loadResults(jobId, activeCategory, page + 1)}><IconChevron dir="right" /></button>
              </div>
            </section>

            <div className="bottom-links">
              <button className="btn-link" onClick={handleViewLog}><IconList /> Lihat Log</button>
              <span className="dim-text">Klik baris tabel untuk lihat detail transaksi</span>
            </div>
          </>
        )}
      </main>

      <SidePanel
        panel={panel}
        onClose={() => setPanel(null)}
        historyList={historyList}
        activeJobId={jobId}
        onSelectHistory={handleSelectHistory}
        onClearHistory={handleClearHistory}
      />

      {showLogModal && (
        <div className="modal-overlay" onClick={() => setShowLogModal(false)}>
          <div className="modal-box" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header"><h3>Log Aktivitas</h3><button onClick={() => setShowLogModal(false)}>×</button></div>
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

      <style jsx global>{`
        html, body { margin: 0; padding: 0; }
        .app {
          font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
          min-height: 100vh;
          background: #f6f7fb;
          color: #14181f;
        }

        /* ---------- topbar ---------- */
        .topbar { background: #ffffff; border-bottom: 1px solid #e6e9f0; }
        .topbar-inner {
          max-width: 1240px; margin: 0 auto; padding: 16px 28px;
          display: flex; justify-content: space-between; align-items: center;
        }
        .brand { display: flex; align-items: center; gap: 14px; }
        .logo-badge {
          background: linear-gradient(135deg, #f2711c, #d85c10);
          color: white; font-weight: 700; padding: 9px 14px; border-radius: 10px;
          font-size: 15px; letter-spacing: 0.5px;
          box-shadow: 0 8px 18px -8px rgba(242, 113, 28, 0.55);
        }
        .brand h1 { margin: 0; font-size: 17px; font-weight: 700; color: #0d1425; }
        .brand p { margin: 2px 0 0; font-size: 12.5px; color: #8188a1; }
        .topbar-actions { display: flex; align-items: center; gap: 10px; }
        .count-pill {
          background: #f2711c; color: white; font-size: 10.5px; font-weight: 700;
          padding: 1px 6px; border-radius: 999px; margin-left: 6px;
        }

        /* ---------- buttons ---------- */
        .btn-outline, .btn-primary-outline, .btn-outline-danger, .btn-chip, .btn-link, .btn-link-danger {
          display: inline-flex; align-items: center; gap: 6px; cursor: pointer;
          font-family: inherit; font-weight: 600;
        }
        .btn-outline {
          background: #ffffff; color: #33394a; border: 1px solid #e0e3ec;
          padding: 8px 14px; border-radius: 9px; font-size: 12.5px; transition: background 0.15s, border-color 0.15s;
          text-decoration: none;
        }
        .btn-outline:hover { background: #f6f7fb; border-color: #cfd4e0; }
        .btn-primary-outline {
          background: #fff3ea; color: #d85c10; border: 1px solid #f8caa0;
          padding: 8px 14px; border-radius: 9px; font-size: 12.5px;
        }
        .btn-primary-outline:hover { background: #fde8d6; }
        .btn-outline-danger {
          background: #fef2f2; color: #dc2626; border: 1px solid #fecaca;
          padding: 10px 20px; border-radius: 9px; font-size: 13.5px;
        }
        .btn-outline-danger:hover { background: #fee2e2; }
        .btn-primary {
          background: linear-gradient(135deg, #f2711c, #d85c10); color: white; border: none;
          padding: 10px 22px; border-radius: 9px; font-size: 13.5px;
          box-shadow: 0 8px 18px -8px rgba(242, 113, 28, 0.5); transition: filter 0.15s, transform 0.1s;
        }
        .btn-primary:hover:not(:disabled) { filter: brightness(1.06); }
        .btn-primary:active:not(:disabled) { transform: translateY(1px); }
        .btn-primary:disabled { background: #d8dbe3; color: #9298a8; box-shadow: none; cursor: not-allowed; }
        .btn-chip:disabled, .btn-link:disabled { opacity: 0.45; cursor: not-allowed; }
        .btn-chip {
          background: #f6f7fb; color: #33394a; border: 1px solid #e0e3ec;
          padding: 7px 12px; border-radius: 7px; font-size: 12px; white-space: nowrap;
        }
        .btn-chip:hover:not(:disabled) { background: #eef0f5; }
        .btn-link {
          background: none; border: none; color: #f2711c; padding: 0; font-size: 12.5px;
        }
        .btn-link-danger {
          background: none; border: none; color: #dc2626; padding: 10px 0 0; font-size: 12.5px; cursor: pointer;
        }
        .btn-page {
          background: #ffffff; color: #33394a; border: 1px solid #e0e3ec; border-radius: 7px;
          width: 30px; height: 30px; display: inline-flex; align-items: center; justify-content: center; cursor: pointer;
        }
        .btn-page:disabled { opacity: 0.35; cursor: not-allowed; }
        .btn-page:hover:not(:disabled) { background: #f6f7fb; }

        /* ---------- layout ---------- */
        .main { max-width: 1240px; margin: 0 auto; padding: 28px 28px 60px; }
        .alert {
          background: #fef2f2; border: 1px solid #fecaca; color: #b91c1c;
          padding: 12px 16px; border-radius: 10px; margin-bottom: 20px; font-size: 13.5px;
        }
        .card {
          background: #ffffff; border-radius: 18px; border: 1px solid #e6e9f0;
          box-shadow: 0 1px 2px rgba(16,24,40,0.04), 0 1px 3px rgba(16,24,40,0.05);
        }
        .card-head { padding: 22px 26px 6px; }
        .card-head h2 { margin: 0 0 4px; font-size: 17px; color: #0d1425; }
        .card-head p { margin: 0; color: #8188a1; font-size: 13px; }

        /* ---------- upload ---------- */
        .upload-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; padding: 18px 26px; }
        .file-slot { border: 1px solid #e6e9f0; border-radius: 14px; padding: 16px; transition: border-color 0.15s; }
        .file-slot-loaded { border-color: #b9e4cf; background: #f7fdfa; }
        .file-slot-head { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
        .file-slot-icon { width: 42px; height: 42px; border-radius: 10px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
        .icon-blue { background: #eef4ff; color: #3b6fe0; }
        .icon-orange { background: #fff3ea; color: #f2711c; }
        .file-slot-label { font-weight: 600; font-size: 13.5px; color: #14181f; }
        .file-slot-hint { font-size: 11.5px; color: #8188a1; margin-top: 1px; }
        .file-slot-row { display: flex; gap: 8px; margin-bottom: 10px; }
        .file-slot-row input[type="text"] {
          flex: 1; min-width: 0; border: 1px solid #e0e3ec; border-radius: 7px; padding: 8px 10px;
          font-size: 12.5px; color: #33394a; background: #fbfbfd; cursor: pointer;
        }
        .file-slot-footer { display: flex; justify-content: space-between; align-items: center; }
        .file-slot-count { font-size: 11.5px; color: #6b7280; font-weight: 600; }
        .file-slot-actions { display: flex; align-items: center; gap: 12px; }
        .upload-footer {
          display: flex; justify-content: space-between; align-items: center;
          padding: 16px 26px 24px; border-top: 1px solid #f0f1f5; margin-top: 4px;
        }
        .status-line { display: flex; align-items: center; gap: 9px; font-size: 13px; color: #4b5165; }
        .status-dot { width: 8px; height: 8px; border-radius: 50%; background: #cbd2e0; }
        .status-dot.status-processing { background: #f2711c; }
        .status-dot.status-error { background: #dc2626; }

        /* ---------- kpi ---------- */
        .kpi-row { display: grid; grid-template-columns: repeat(6, 1fr); gap: 14px; margin-bottom: 16px; }
        .kpi-card {
          background: #ffffff; border: 1px solid #e6e9f0; border-radius: 14px; padding: 16px;
          display: flex; flex-direction: column; gap: 6px; text-align: left; cursor: default; font-family: inherit;
          box-shadow: 0 1px 2px rgba(16,24,40,0.03);
        }
        button.kpi-card { cursor: pointer; transition: border-color 0.15s; }
        button.kpi-card:hover { border-color: #cfd4e0; }
        .kpi-card-active { border-width: 1.5px; }
        .kpi-card-plain { background: #fafbfc; }
        .kpi-label { font-size: 12px; font-weight: 600; color: #6b7280; }
        .kpi-value { font-size: 21px; font-weight: 700; color: #14181f; }

        .nominal-strip {
          display: flex; gap: 28px; background: #ffffff; border: 1px solid #e6e9f0; border-radius: 14px;
          padding: 14px 22px; margin-bottom: 20px; flex-wrap: wrap;
        }
        .nominal-strip div { display: flex; flex-direction: column; gap: 3px; }
        .nominal-strip span { font-size: 11.5px; color: #8188a1; font-weight: 600; }
        .nominal-strip strong { font-size: 14.5px; color: #14181f; }
        .text-red { color: #dc2626 !important; }

        /* ---------- results ---------- */
        .results-header {
          display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px;
          padding: 18px 24px;
        }
        .filter-pills { display: flex; gap: 8px; flex-wrap: wrap; }
        .pill-btn {
          background: #f6f7fb; color: #4b5165; border: 1px solid transparent; border-radius: 999px;
          padding: 7px 14px; font-size: 12px; font-weight: 600; cursor: pointer; font-family: inherit;
        }
        .pill-btn:hover { background: #eef0f5; }
        .pill-btn-active { background: #ffffff; border-color: #e0e3ec; }
        .results-actions { display: flex; gap: 8px; }

        .table-wrap { overflow-x: auto; border-top: 1px solid #f0f1f5; }
        .data-table { width: 100%; border-collapse: collapse; font-size: 12.5px; white-space: nowrap; }
        .data-table th {
          text-align: left; padding: 10px 16px; font-weight: 700; color: #6b7280; font-size: 11px;
          text-transform: uppercase; letter-spacing: 0.03em; border-bottom: 1px solid #f0f1f5; background: #fafbfc;
        }
        .data-table td { padding: 11px 16px; border-bottom: 1px solid #f5f6f9; color: #33394a; }
        .data-table tbody tr { cursor: pointer; transition: background 0.1s; }
        .data-table tbody tr:hover { background: #fafbfc; }
        .font-mono { font-family: ui-monospace, "SF Mono", Menlo, monospace; font-size: 12px; }
        .key-pill { background: #f0f1f5; border-radius: 5px; padding: 2px 7px; font-size: 11.5px; font-weight: 600; }
        .empty-state { text-align: center; color: #8188a1; padding: 34px !important; cursor: default; }
        .badge {
          display: inline-block; font-size: 11px; font-weight: 700; padding: 3px 9px;
          border-radius: 999px; border: 1px solid; white-space: nowrap;
        }
        .pagination {
          display: flex; align-items: center; justify-content: center; gap: 14px;
          padding: 16px; font-size: 12.5px; color: #6b7280; border-top: 1px solid #f0f1f5;
        }

        .bottom-links { display: flex; align-items: center; gap: 16px; margin-top: 18px; }
        .dim-text { color: #8188a1; font-size: 12.5px; }

        /* ---------- side panel ---------- */
        .panel-overlay {
          position: fixed; inset: 0; background: rgba(13, 20, 37, 0.35); z-index: 40;
          display: flex; justify-content: flex-end;
          animation: fadeIn 0.15s ease-out;
        }
        @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
        .side-panel {
          width: 380px; max-width: 92vw; height: 100%; background: #12182b; color: #eef2fa;
          padding: 24px; overflow-y: auto; box-shadow: -12px 0 30px rgba(0,0,0,0.25);
          animation: slideIn 0.2s ease-out;
        }
        @keyframes slideIn { from { transform: translateX(24px); opacity: 0; } to { transform: translateX(0); opacity: 1; } }
        .side-panel-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
        .side-panel-head h3 { margin: 0; font-size: 16px; }
        .panel-close { background: none; border: none; color: #8b93ab; font-size: 22px; cursor: pointer; line-height: 1; }
        .panel-close:hover { color: #eef2fa; }

        .history-list-v { display: flex; flex-direction: column; gap: 10px; margin-bottom: 8px; }
        .history-row { border: 1px solid #232b45; border-radius: 12px; padding: 12px 14px; cursor: pointer; transition: border-color 0.15s, background 0.15s; }
        .history-row:hover { border-color: #37507e; background: #161f38; }
        .history-row-active { border-color: #f2711c; background: #1c1712; }
        .history-row-top { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 8px; gap: 10px; }
        .history-row-top strong { font-size: 13px; }
        .history-row-top span { font-size: 11px; color: #6f7996; white-space: nowrap; }
        .history-row-stats { display: flex; gap: 8px; }
        .pill { font-size: 10.5px; font-weight: 700; padding: 3px 8px; border-radius: 999px; }
        .pill-green { background: rgba(52, 211, 153, 0.15); color: #34d399; }
        .pill-red { background: rgba(251, 113, 133, 0.15); color: #fb7185; }

        .detail-amounts { display: flex; gap: 20px; margin-bottom: 14px; }
        .detail-amounts div { display: flex; flex-direction: column; gap: 4px; }
        .detail-amounts span { font-size: 11px; color: #8b93ab; }
        .detail-amounts strong { font-size: 19px; }
        .detail-rows { margin-top: 18px; display: flex; flex-direction: column; }
        .detail-row {
          display: flex; justify-content: space-between; gap: 12px; padding: 10px 0;
          border-bottom: 1px solid #1f2740; font-size: 12.5px;
        }
        .detail-row span { color: #8b93ab; }
        .detail-row strong { color: #eef2fa; text-align: right; font-weight: 600; }
        .detail-note { margin-top: 16px; }
        .detail-note span { display: block; font-size: 11px; color: #8b93ab; margin-bottom: 5px; }
        .detail-note p { margin: 0; font-size: 12.5px; line-height: 1.5; }

        /* ---------- modals ---------- */
        .modal-overlay {
          position: fixed; inset: 0; background: rgba(13,20,37,0.4); z-index: 50;
          display: flex; align-items: center; justify-content: center; padding: 20px;
        }
        .modal-box {
          background: #ffffff; border-radius: 16px; max-width: 480px; width: 100%;
          max-height: 80vh; display: flex; flex-direction: column; overflow: hidden;
          box-shadow: 0 20px 40px rgba(0,0,0,0.2);
        }
        .modal-wide { max-width: 720px; }
        .modal-header {
          display: flex; justify-content: space-between; align-items: center;
          padding: 16px 20px; border-bottom: 1px solid #f0f1f5;
        }
        .modal-header h3 { margin: 0; font-size: 15px; }
        .modal-header button { background: none; border: none; font-size: 20px; cursor: pointer; color: #8188a1; }
        .modal-body { padding: 16px 20px; overflow-y: auto; }
        .log-entry { display: flex; gap: 10px; padding: 7px 0; font-size: 12.5px; border-bottom: 1px solid #f5f6f9; }
        .log-time { color: #8188a1; font-family: ui-monospace, monospace; font-size: 11.5px; flex-shrink: 0; }
        .preview-pre {
          margin: 0; font-size: 11.5px; font-family: ui-monospace, "SF Mono", Menlo, monospace;
          white-space: pre-wrap; word-break: break-all; color: #33394a;
        }

        @media (max-width: 900px) {
          .upload-grid { grid-template-columns: 1fr; }
          .kpi-row { grid-template-columns: repeat(2, 1fr); }
        }
      `}</style>
    </div>
  );
}
