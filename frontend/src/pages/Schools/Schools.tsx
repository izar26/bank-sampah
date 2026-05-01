import { useEffect, useState } from "react";
import PageMeta from "../../components/common/PageMeta";
import { schoolAPI, School } from "../../api";

export default function Schools() {
  const [schools, setSchools] = useState<School[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [newSchool, setNewSchool] = useState({ name: "", code: "", callback_url: "" });
  const [createdCreds, setCreatedCreds] = useState<{ api_key: string; api_secret: string } | null>(null);

  const loadSchools = async () => {
    try {
      const res = await schoolAPI.list();
      if (res.success && res.data) setSchools(res.data);
    } finally { setLoading(false); }
  };

  useEffect(() => { loadSchools(); }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await schoolAPI.create(newSchool);
      if (res.success && res.data) {
        setCreatedCreds({ api_key: res.data.api_key, api_secret: res.data.api_secret || "" });
        setNewSchool({ name: "", code: "", callback_url: "" });
        loadSchools();
      } else alert(res.message);
    } catch (err) { console.error(err); }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("Apakah Anda yakin ingin menghapus sekolah ini? Data SI yang terhubung juga mungkin terpengaruh.")) return;
    try {
      const res = await schoolAPI.delete(id);
      if (res.success) loadSchools();
      else alert(res.message);
    } catch (err) { console.error(err); }
  };

  const handleToggleStatus = async (id: string, currentStatus: boolean) => {
    if (!window.confirm(`Yakin ingin ${currentStatus ? 'menonaktifkan' : 'mengaktifkan'} sekolah ini?`)) return;
    try {
      const res = await schoolAPI.update(id, { is_active: !currentStatus });
      if (res.success) loadSchools();
      else alert(res.message);
    } catch (err) { console.error(err); }
  };

  return (
    <>
      <PageMeta title="Sekolah | Bank Sampah" description="Manajemen Sekolah" />
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-800 dark:text-white/90">Sekolah</h1>
            <p className="text-sm text-gray-500">Kelola sekolah yang terdaftar</p>
          </div>
          <button onClick={() => { setShowCreate(!showCreate); setCreatedCreds(null); }} className="px-4 py-2 text-sm font-medium text-white rounded-lg bg-brand-500 hover:bg-brand-600">
            + Tambah Sekolah
          </button>
        </div>

        {showCreate && (
          <div className="p-6 bg-white border border-gray-200 rounded-2xl dark:bg-gray-800 dark:border-gray-700">
            <h3 className="mb-4 text-lg font-semibold text-gray-800 dark:text-white/90">Daftarkan Sekolah Baru</h3>
            <form onSubmit={handleCreate} className="grid gap-4 sm:grid-cols-3">
              <input value={newSchool.name} onChange={e => setNewSchool({ ...newSchool, name: e.target.value })} placeholder="Nama Sekolah" required className="px-4 py-2 text-sm border border-gray-200 rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white focus:ring-2 focus:ring-brand-500 focus:outline-none" />
              <input value={newSchool.code} onChange={e => setNewSchool({ ...newSchool, code: e.target.value })} placeholder="Kode (unik)" required className="px-4 py-2 text-sm border border-gray-200 rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white focus:ring-2 focus:ring-brand-500 focus:outline-none" />
              <input value={newSchool.callback_url} onChange={e => setNewSchool({ ...newSchool, callback_url: e.target.value })} placeholder="Callback URL (opsional)" className="px-4 py-2 text-sm border border-gray-200 rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white focus:ring-2 focus:ring-brand-500 focus:outline-none" />
              <button type="submit" className="px-4 py-2 text-sm font-medium text-white rounded-lg bg-brand-500 hover:bg-brand-600 sm:col-span-3">Simpan</button>
            </form>
            {createdCreds && (
              <div className="p-4 mt-4 border border-green-200 rounded-xl bg-green-50 dark:bg-green-500/10 dark:border-green-500/30">
                <p className="mb-2 text-sm font-semibold text-green-700 dark:text-green-400">⚠️ Simpan kredensial ini! API Secret hanya ditampilkan sekali.</p>
                <p className="text-xs text-gray-600 dark:text-gray-300"><strong>API Key:</strong> <code className="px-2 py-1 bg-gray-100 rounded dark:bg-gray-700">{createdCreds.api_key}</code></p>
                <p className="mt-1 text-xs text-gray-600 dark:text-gray-300"><strong>API Secret:</strong> <code className="px-2 py-1 bg-gray-100 rounded dark:bg-gray-700">{createdCreds.api_secret}</code></p>
              </div>
            )}
          </div>
        )}

        <div className="overflow-hidden bg-white border border-gray-200 rounded-2xl dark:bg-gray-800 dark:border-gray-700">
          {loading ? (
            <div className="flex items-center justify-center h-48"><div className="w-8 h-8 border-4 border-brand-500 border-t-transparent rounded-full animate-spin" /></div>
          ) : schools.length === 0 ? (
            <div className="p-8 text-center text-gray-500">Belum ada sekolah terdaftar</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead><tr className="border-b border-gray-200 dark:border-gray-700">
                  {["Nama", "Kode", "Kredensial API", "Callback URL", "Status", "Aksi"].map(h => <th key={h} className="px-5 py-4 text-xs font-medium text-left text-gray-500 uppercase">{h}</th>)}
                </tr></thead>
                <tbody>{schools.map(s => (
                  <tr key={s.id} className="border-b border-gray-100 dark:border-gray-700/50">
                    <td className="px-5 py-4 text-sm font-medium text-gray-800 dark:text-white/90">{s.name}</td>
                    <td className="px-5 py-4 text-sm text-gray-600 dark:text-gray-300">{s.code}</td>
                    <td className="px-5 py-4">
                      <div className="flex flex-col space-y-2">
                        <div className="flex items-center space-x-2">
                          <span className="text-[10px] font-semibold text-gray-400 uppercase w-12">Key</span>
                          <code className="px-2 py-1 text-xs font-mono text-gray-600 bg-gray-100 rounded dark:bg-gray-800 dark:text-gray-300">{s.api_key.substring(0, 10)}...</code>
                          <button onClick={() => { navigator.clipboard.writeText(s.api_key); alert('API Key berhasil disalin!'); }} className="text-gray-400 hover:text-brand-500 transition-colors" title="Salin API Key">
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path></svg>
                          </button>
                        </div>
                        <div className="flex items-center space-x-2">
                          <span className="text-[10px] font-semibold text-gray-400 uppercase w-12">Secret</span>
                          <code className="px-2 py-1 text-xs font-mono text-gray-600 bg-gray-100 rounded dark:bg-gray-800 dark:text-gray-300">••••••••••</code>
                          <button onClick={() => { navigator.clipboard.writeText(s.api_secret || ''); alert('API Secret berhasil disalin!'); }} className="text-gray-400 hover:text-brand-500 transition-colors" title="Salin API Secret">
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path></svg>
                          </button>
                        </div>
                      </div>
                    </td>
                    <td className="px-5 py-4 text-sm text-gray-500 dark:text-gray-400">{s.callback_url || "-"}</td>
                    <td className="px-5 py-4"><span className={`px-2.5 py-1 text-xs font-medium rounded-full ${s.is_active ? "bg-green-100 text-green-600 dark:bg-green-500/15 dark:text-green-400" : "bg-red-100 text-red-600 dark:bg-red-500/15 dark:text-red-400"}`}>{s.is_active ? "Aktif" : "Nonaktif"}</span></td>
                    <td className="px-5 py-4 text-sm font-medium">
                      <div className="flex items-center space-x-3">
                        <button onClick={() => handleToggleStatus(s.id, s.is_active)} className={`${s.is_active ? 'text-yellow-500 hover:text-yellow-600' : 'text-green-500 hover:text-green-600'} transition-colors`}>
                          {s.is_active ? "Nonaktifkan" : "Aktifkan"}
                        </button>
                        <button onClick={() => handleDelete(s.id)} className="text-red-500 transition-colors hover:text-red-600">
                          Hapus
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}</tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </>
  );
}
