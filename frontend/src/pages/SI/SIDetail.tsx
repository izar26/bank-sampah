import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router";
import PageMeta from "../../components/common/PageMeta";
import { siAPI, SIDocument } from "../../api";

export default function SIDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [doc, setDoc] = useState<SIDocument | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [rejectReason, setRejectReason] = useState("");
  const [showReject, setShowReject] = useState(false);

  useEffect(() => { if (id) loadDetail(); }, [id]);

  const loadDetail = async () => {
    try {
      const res = await siAPI.getDetail(id!);
      if (res.success && res.data) setDoc(res.data);
    } catch (err) { console.error(err); }
    finally { setLoading(false); }
  };

  const handleAction = async (action: "verify" | "approve" | "disburse") => {
    setActionLoading(true);
    try {
      const fn = action === "verify" ? siAPI.verify : action === "approve" ? siAPI.approve : siAPI.disburse;
      const res = await fn(id!);
      if (res.success) loadDetail(); else alert(res.message);
    } finally { setActionLoading(false); }
  };

  const handleReject = async () => {
    setActionLoading(true);
    try {
      const res = await siAPI.reject(id!, rejectReason);
      if (res.success) { setShowReject(false); loadDetail(); } else alert(res.message);
    } finally { setActionLoading(false); }
  };

  const fmt = (n: number) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(n);
  const fmtDate = (d: string | null) => d ? new Date(d).toLocaleDateString("id-ID", { day: "2-digit", month: "long", year: "numeric", hour: "2-digit", minute: "2-digit" }) : "-";

  if (loading) return <div className="flex items-center justify-center h-64"><div className="w-8 h-8 border-4 border-brand-500 border-t-transparent rounded-full animate-spin" /></div>;
  if (!doc) return <div className="p-8 text-center text-gray-500">Dokumen tidak ditemukan</div>;

  return (
    <>
      <PageMeta title={`SI ${doc.si_number}`} description="Detail SI" />
      <div className="space-y-6">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <button onClick={() => navigate("/si")} className="mb-2 text-sm text-gray-500 hover:text-brand-500">← Kembali</button>
            <h1 className="text-2xl font-bold text-gray-800 dark:text-white/90">SI #{doc.si_number}</h1>
            <p className="text-sm text-gray-500">{doc.school?.name || "-"}</p>
          </div>
          <div className="flex gap-2 flex-wrap">
            {(doc.status === "PENDING" || doc.status === "PROCESSING") && <button onClick={() => handleAction("verify")} disabled={actionLoading} className="px-4 py-2 text-sm font-medium text-white bg-cyan-500 rounded-lg hover:bg-cyan-600 disabled:opacity-50">✓ Verifikasi</button>}
            {doc.status === "VERIFIED" && <button onClick={() => handleAction("approve")} disabled={actionLoading} className="px-4 py-2 text-sm font-medium text-white bg-indigo-500 rounded-lg hover:bg-indigo-600 disabled:opacity-50">✓ Approve</button>}
            {doc.status === "APPROVED" && <button onClick={() => handleAction("disburse")} disabled={actionLoading} className="px-4 py-2 text-sm font-medium text-white bg-green-500 rounded-lg hover:bg-green-600 disabled:opacity-50">💰 Cairkan</button>}
            {doc.status !== "DISBURSED" && doc.status !== "REJECTED" && <button onClick={() => setShowReject(!showReject)} className="px-4 py-2 text-sm font-medium text-white bg-red-500 rounded-lg hover:bg-red-600">✕ Tolak</button>}
          </div>
        </div>

        {showReject && (
          <div className="p-4 bg-red-50 border border-red-200 rounded-xl dark:bg-red-500/10 dark:border-red-500/30">
            <textarea value={rejectReason} onChange={(e) => setRejectReason(e.target.value)} className="w-full p-3 mb-3 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 dark:text-white" rows={2} placeholder="Alasan penolakan..." />
            <button onClick={handleReject} disabled={actionLoading} className="px-4 py-2 text-sm text-white bg-red-500 rounded-lg mr-2">Konfirmasi</button>
            <button onClick={() => setShowReject(false)} className="px-4 py-2 text-sm bg-gray-100 rounded-lg dark:bg-gray-700 dark:text-gray-300">Batal</button>
          </div>
        )}

        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {[{ l: "Status", v: doc.status }, { l: "Nasabah", v: doc.total_items }, { l: "Total", v: fmt(doc.total_amount) }, { l: "Dibuat", v: fmtDate(doc.created_at) }].map(i => (
            <div key={i.l} className="p-4 bg-white border border-gray-200 rounded-xl dark:bg-gray-800 dark:border-gray-700">
              <p className="text-xs text-gray-500">{i.l}</p>
              <p className="mt-1 text-sm font-semibold text-gray-800 dark:text-white/90">{i.v}</p>
            </div>
          ))}
        </div>

        {doc.items && doc.items.length > 0 && (
          <div className="overflow-hidden bg-white border border-gray-200 rounded-2xl dark:bg-gray-800 dark:border-gray-700">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-lg font-semibold text-gray-800 dark:text-white/90">Nasabah ({doc.items.length})</h2>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead><tr className="border-b border-gray-200 dark:border-gray-700">
                  {["#", "Nama", "ID", "Tipe", "Nominal", "Status"].map(h => <th key={h} className="px-5 py-3 text-xs font-medium text-left text-gray-500 uppercase">{h}</th>)}
                </tr></thead>
                <tbody>{doc.items.map((item, idx) => (
                  <tr key={item.id} className="border-b border-gray-100 dark:border-gray-700/50">
                    <td className="px-5 py-3 text-sm text-gray-500">{idx + 1}</td>
                    <td className="px-5 py-3 text-sm font-medium text-gray-800 dark:text-white/90">{item.nasabah_name}</td>
                    <td className="px-5 py-3 text-sm text-gray-600 dark:text-gray-300">{item.nasabah_identifier}</td>
                    <td className="px-5 py-3"><span className={`px-2 py-0.5 text-xs font-medium rounded-full ${item.nasabah_type === "siswa" ? "bg-blue-100 text-blue-600 dark:bg-blue-500/15 dark:text-blue-400" : "bg-purple-100 text-purple-600 dark:bg-purple-500/15 dark:text-purple-400"}`}>{item.nasabah_type === "siswa" ? "Siswa" : "GTK"}</span></td>
                    <td className="px-5 py-3 text-sm font-medium text-gray-800 dark:text-white/90">{fmt(item.amount)}</td>
                    <td className="px-5 py-3 text-sm text-gray-500">{item.status}</td>
                  </tr>
                ))}</tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </>
  );
}
