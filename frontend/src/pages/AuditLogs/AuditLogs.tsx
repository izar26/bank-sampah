import { useEffect, useState } from "react";
import PageMeta from "../../components/common/PageMeta";
import { auditAPI, AuditLog } from "../../api";

export default function AuditLogs() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => { loadLogs(); }, [page]);

  const loadLogs = async () => {
    setLoading(true);
    try {
      const res = await auditAPI.list(page, 20);
      if (res.success && res.data) {
        setLogs(res.data);
        if (res.meta) setTotalPages(res.meta.total_pages);
      }
    } finally { setLoading(false); }
  };

  const fmtDate = (d: string) => new Date(d).toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" });

  return (
    <>
      <PageMeta title="Audit Log | Bank Sampah" description="Log Aktivitas Sistem" />
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-white/90">Audit Log</h1>
          <p className="text-sm text-gray-500">Riwayat semua aktivitas di sistem</p>
        </div>
        <div className="overflow-hidden bg-white border border-gray-200 rounded-2xl dark:bg-gray-800 dark:border-gray-700">
          {loading ? (
            <div className="flex items-center justify-center h-48"><div className="w-8 h-8 border-4 border-brand-500 border-t-transparent rounded-full animate-spin" /></div>
          ) : logs.length === 0 ? (
            <div className="p-8 text-center text-gray-500">Belum ada aktivitas</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead><tr className="border-b border-gray-200 dark:border-gray-700">
                  {["Waktu", "Admin", "Aksi", "IP Address"].map(h => <th key={h} className="px-5 py-4 text-xs font-medium text-left text-gray-500 uppercase">{h}</th>)}
                </tr></thead>
                <tbody>{logs.map(log => (
                  <tr key={log.id} className="border-b border-gray-100 dark:border-gray-700/50">
                    <td className="px-5 py-3 text-sm text-gray-500">{fmtDate(log.created_at)}</td>
                    <td className="px-5 py-3 text-sm font-medium text-gray-800 dark:text-white/90">{log.admin?.username || "System"}</td>
                    <td className="px-5 py-3 text-sm text-gray-600 dark:text-gray-300 font-mono">{log.action}</td>
                    <td className="px-5 py-3 text-sm text-gray-500">{log.ip_address}</td>
                  </tr>
                ))}</tbody>
              </table>
            </div>
          )}
          {totalPages > 1 && (
            <div className="flex items-center justify-between px-5 py-4 border-t border-gray-200 dark:border-gray-700">
              <button onClick={() => setPage(Math.max(1, page - 1))} disabled={page === 1} className="px-4 py-2 text-sm bg-gray-100 rounded-lg disabled:opacity-50 dark:bg-gray-700 dark:text-gray-300">Sebelumnya</button>
              <span className="text-sm text-gray-500">Halaman {page} dari {totalPages}</span>
              <button onClick={() => setPage(Math.min(totalPages, page + 1))} disabled={page === totalPages} className="px-4 py-2 text-sm bg-gray-100 rounded-lg disabled:opacity-50 dark:bg-gray-700 dark:text-gray-300">Selanjutnya</button>
            </div>
          )}
        </div>
      </div>
    </>
  );
}
