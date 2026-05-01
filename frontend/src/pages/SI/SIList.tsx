import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import PageMeta from "../../components/common/PageMeta";
import { siAPI, SIDocument } from "../../api";

const statusBadge: Record<string, { bg: string; text: string; label: string }> = {
  PENDING: { bg: "bg-yellow-100 dark:bg-yellow-500/15", text: "text-yellow-600 dark:text-yellow-400", label: "Pending" },
  PROCESSING: { bg: "bg-blue-100 dark:bg-blue-500/15", text: "text-blue-600 dark:text-blue-400", label: "Processing" },
  VERIFIED: { bg: "bg-cyan-100 dark:bg-cyan-500/15", text: "text-cyan-600 dark:text-cyan-400", label: "Diverifikasi" },
  APPROVED: { bg: "bg-indigo-100 dark:bg-indigo-500/15", text: "text-indigo-600 dark:text-indigo-400", label: "Di-approve" },
  DISBURSED: { bg: "bg-green-100 dark:bg-green-500/15", text: "text-green-600 dark:text-green-400", label: "Dicairkan" },
  REJECTED: { bg: "bg-red-100 dark:bg-red-500/15", text: "text-red-600 dark:text-red-400", label: "Ditolak" },
};

export default function SIList() {
  const [documents, setDocuments] = useState<SIDocument[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [filterStatus, setFilterStatus] = useState("");
  const navigate = useNavigate();

  const loadDocuments = async () => {
    setLoading(true);
    try {
      const res = await siAPI.list(page, 15, filterStatus);
      if (res.success && res.data) {
        setDocuments(res.data);
        if (res.meta) setTotalPages(res.meta.total_pages);
      }
    } catch (err) {
      console.error("Failed to load SI documents:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    let active = true;
    setTimeout(() => {
      if (active) loadDocuments();
    }, 0);
    return () => { active = false; };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, filterStatus]);

  const formatCurrency = (amount: number) =>
    new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(amount);

  const formatDate = (dateStr: string) =>
    new Date(dateStr).toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" });

  return (
    <>
      <PageMeta title="Surat Instruksi | Bank Sampah" description="Daftar Surat Instruksi" />
      <div className="space-y-6">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-800 dark:text-white/90">Surat Instruksi</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400">Kelola semua Surat Instruksi dari sekolah</p>
          </div>
          <select
            value={filterStatus}
            onChange={(e) => { setFilterStatus(e.target.value); setPage(1); }}
            className="px-4 py-2 text-sm bg-white border border-gray-200 rounded-lg dark:bg-gray-800 dark:border-gray-700 dark:text-white/90 focus:outline-none focus:ring-2 focus:ring-brand-500"
          >
            <option value="">Semua Status</option>
            <option value="PENDING">Pending</option>
            <option value="VERIFIED">Diverifikasi</option>
            <option value="APPROVED">Di-approve</option>
            <option value="DISBURSED">Dicairkan</option>
            <option value="REJECTED">Ditolak</option>
          </select>
        </div>

        <div className="overflow-hidden bg-white border border-gray-200 rounded-2xl dark:bg-gray-800 dark:border-gray-700">
          {loading ? (
            <div className="flex items-center justify-center h-48">
              <div className="w-8 h-8 border-4 border-brand-500 border-t-transparent rounded-full animate-spin"></div>
            </div>
          ) : documents.length === 0 ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">
              Belum ada Surat Instruksi
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200 dark:border-gray-700">
                    <th className="px-5 py-4 text-xs font-medium text-left text-gray-500 uppercase dark:text-gray-400">No. SI</th>
                    <th className="px-5 py-4 text-xs font-medium text-left text-gray-500 uppercase dark:text-gray-400">Sekolah</th>
                    <th className="px-5 py-4 text-xs font-medium text-left text-gray-500 uppercase dark:text-gray-400">Nasabah</th>
                    <th className="px-5 py-4 text-xs font-medium text-left text-gray-500 uppercase dark:text-gray-400">Total</th>
                    <th className="px-5 py-4 text-xs font-medium text-left text-gray-500 uppercase dark:text-gray-400">Status</th>
                    <th className="px-5 py-4 text-xs font-medium text-left text-gray-500 uppercase dark:text-gray-400">Tanggal</th>
                  </tr>
                </thead>
                <tbody>
                  {documents.map((doc) => {
                    const badge = statusBadge[doc.status] || statusBadge.PENDING;
                    return (
                      <tr
                        key={doc.id}
                        onClick={() => navigate(`/si/${doc.id}`)}
                        className="border-b border-gray-100 cursor-pointer hover:bg-gray-50 dark:border-gray-700/50 dark:hover:bg-gray-700/30 transition-colors"
                      >
                        <td className="px-5 py-4 text-sm font-medium text-gray-800 dark:text-white/90">{doc.si_number}</td>
                        <td className="px-5 py-4 text-sm text-gray-600 dark:text-gray-300">{doc.school?.name || "-"}</td>
                        <td className="px-5 py-4 text-sm text-gray-600 dark:text-gray-300">{doc.total_items.toLocaleString("id-ID")}</td>
                        <td className="px-5 py-4 text-sm font-medium text-gray-800 dark:text-white/90">{formatCurrency(doc.total_amount)}</td>
                        <td className="px-5 py-4">
                          <span className={`inline-flex px-2.5 py-1 text-xs font-medium rounded-full ${badge.bg} ${badge.text}`}>
                            {badge.label}
                          </span>
                        </td>
                        <td className="px-5 py-4 text-sm text-gray-500 dark:text-gray-400">{formatDate(doc.created_at)}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between px-5 py-4 border-t border-gray-200 dark:border-gray-700">
              <button
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page === 1}
                className="px-4 py-2 text-sm font-medium text-gray-600 bg-gray-100 rounded-lg disabled:opacity-50 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
              >
                Sebelumnya
              </button>
              <span className="text-sm text-gray-500 dark:text-gray-400">
                Halaman {page} dari {totalPages}
              </span>
              <button
                onClick={() => setPage(Math.min(totalPages, page + 1))}
                disabled={page === totalPages}
                className="px-4 py-2 text-sm font-medium text-gray-600 bg-gray-100 rounded-lg disabled:opacity-50 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
              >
                Selanjutnya
              </button>
            </div>
          )}
        </div>
      </div>
    </>
  );
}
