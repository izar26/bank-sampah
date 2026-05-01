import { useEffect, useState } from "react";
import PageMeta from "../../components/common/PageMeta";
import { dashboardAPI } from "../../api";

interface Stats {
  total_schools: number;
  total_si: number;
  pending_si: number;
  processing_si: number;
  verified_si: number;
  approved_si: number;
  disbursed_si: number;
  rejected_si: number;
  total_disbursed: number;
  total_nasabah: number;
}

export default function Home() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      const res = await dashboardAPI.getStats();
      if (res.success && res.data) {
        setStats(res.data);
      }
    } catch (err) {
      console.error("Failed to load dashboard stats:", err);
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(amount);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-brand-500 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  const metricCards = [
    {
      title: "Total Sekolah",
      value: stats?.total_schools || 0,
      icon: (
        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
        </svg>
      ),
      color: "bg-blue-500/10 text-blue-600 dark:text-blue-400",
    },
    {
      title: "Total SI",
      value: stats?.total_si || 0,
      icon: (
        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
      ),
      color: "bg-purple-500/10 text-purple-600 dark:text-purple-400",
    },
    {
      title: "Total Nasabah",
      value: stats?.total_nasabah || 0,
      icon: (
        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
      ),
      color: "bg-green-500/10 text-green-600 dark:text-green-400",
    },
    {
      title: "Total Dicairkan",
      value: formatCurrency(stats?.total_disbursed || 0),
      icon: (
        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      ),
      color: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
    },
  ];

  const statusCards = [
    { label: "Pending", value: stats?.pending_si || 0, color: "bg-yellow-400" },
    { label: "Diverifikasi", value: stats?.verified_si || 0, color: "bg-blue-400" },
    { label: "Di-approve", value: stats?.approved_si || 0, color: "bg-indigo-400" },
    { label: "Dicairkan", value: stats?.disbursed_si || 0, color: "bg-green-400" },
    { label: "Ditolak", value: stats?.rejected_si || 0, color: "bg-red-400" },
  ];

  return (
    <>
      <PageMeta
        title="Dashboard | Bank Sampah"
        description="Dashboard Sistem Manajemen Bank Sampah"
      />
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-white/90">
            Dashboard
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Ringkasan aktivitas sistem Bank Sampah
          </p>
        </div>

        {/* Metric Cards */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {metricCards.map((card) => (
            <div
              key={card.title}
              className="p-5 bg-white border border-gray-200 rounded-2xl dark:bg-gray-800 dark:border-gray-700"
            >
              <div className="flex items-center gap-4">
                <div className={`flex items-center justify-center w-12 h-12 rounded-xl ${card.color}`}>
                  {card.icon}
                </div>
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">{card.title}</p>
                  <h3 className="text-xl font-bold text-gray-800 dark:text-white/90">
                    {typeof card.value === "number" ? card.value.toLocaleString("id-ID") : card.value}
                  </h3>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* SI Status Overview */}
        <div className="p-6 bg-white border border-gray-200 rounded-2xl dark:bg-gray-800 dark:border-gray-700">
          <h2 className="mb-4 text-lg font-semibold text-gray-800 dark:text-white/90">
            Status Surat Instruksi
          </h2>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
            {statusCards.map((item) => (
              <div key={item.label} className="p-4 rounded-xl bg-gray-50 dark:bg-gray-700/50">
                <div className="flex items-center gap-2 mb-2">
                  <div className={`w-3 h-3 rounded-full ${item.color}`}></div>
                  <span className="text-sm text-gray-500 dark:text-gray-400">{item.label}</span>
                </div>
                <p className="text-2xl font-bold text-gray-800 dark:text-white/90">
                  {item.value.toLocaleString("id-ID")}
                </p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </>
  );
}
