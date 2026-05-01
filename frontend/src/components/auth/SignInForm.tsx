import { useState } from "react";
import { useNavigate } from "react-router";
import { EyeCloseIcon, EyeIcon } from "../../icons";
import Label from "../form/Label";
import Input from "../form/input/InputField";
import Button from "../ui/button/Button";
import { authAPI, setTokens } from "../../api";

export default function SignInForm() {
  const [showPassword, setShowPassword] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const res = await authAPI.login(username, password);
      if (res.success && res.data) {
        setTokens(res.data.access_token, res.data.refresh_token);
        navigate("/");
      } else {
        setError(res.message || "Login gagal");
      }
    } catch {
      setError("Tidak dapat terhubung ke server");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col flex-1">
      <div className="flex flex-col justify-center flex-1 w-full max-w-md mx-auto">
        <div>
          <div className="mb-5 sm:mb-8">
            <div className="flex items-center gap-3 mb-4">
              <div className="flex items-center justify-center w-12 h-12 rounded-xl bg-brand-500/10">
                <svg className="w-7 h-7 text-brand-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-800 dark:text-white/90">
                  Bank Sampah
                </h1>
                <p className="text-xs text-gray-400 dark:text-gray-500">
                  Sistem Manajemen Pencairan Dana
                </p>
              </div>
            </div>
            <h2 className="mb-2 font-semibold text-gray-800 text-title-sm dark:text-white/90">
              Masuk ke Dashboard
            </h2>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Masukkan username dan password Anda
            </p>
          </div>
          <div>
            {error && (
              <div className="p-3 mb-4 text-sm text-red-600 rounded-lg bg-red-50 dark:bg-red-500/10 dark:text-red-400">
                {error}
              </div>
            )}
            <form onSubmit={handleSubmit}>
              <div className="space-y-5">
                <div>
                  <Label>
                    Username <span className="text-error-500">*</span>
                  </Label>
                  <Input
                    placeholder="admin"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                  />
                </div>
                <div>
                  <Label>
                    Password <span className="text-error-500">*</span>
                  </Label>
                  <div className="relative">
                    <Input
                      type={showPassword ? "text" : "password"}
                      placeholder="Masukkan password"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                    />
                    <span
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute z-30 -translate-y-1/2 cursor-pointer right-4 top-1/2"
                    >
                      {showPassword ? (
                        <EyeIcon className="fill-gray-500 dark:fill-gray-400 size-5" />
                      ) : (
                        <EyeCloseIcon className="fill-gray-500 dark:fill-gray-400 size-5" />
                      )}
                    </span>
                  </div>
                </div>
                <div>
                  <Button className="w-full" size="sm" disabled={loading}>
                    {loading ? "Memproses..." : "Masuk"}
                  </Button>
                </div>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
