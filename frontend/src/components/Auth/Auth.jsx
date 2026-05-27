import { useState } from "react";
import { supabase } from "../../lib/supabase";

export default function Auth() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");
  const [isLogin, setIsLogin] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async () => {
    setLoading(true);
    setError("");
    if (isLogin) {
      const { error } = await supabase.auth.signInWithPassword({ email, password });
      if (error) setError(error.message);
    } else {
      const { data, error } = await supabase.auth.signUp({ email, password });
      if (error) {
        setError(error.message);
      } else if (data.user) {
        await supabase.from("profiles").insert({ id: data.user.id, username, is_online: false });
      }
    }
    setLoading(false);
  };

  return (
    <div className="auth-bg">
      <div className="wood-grain-overlay" />
      <div className="auth-card glass-panel">
        <div className="auth-logo">
          <span className="auth-logo-icon">🪵</span>
          <h1 className="auth-title">Timber</h1>
          <p className="auth-subtitle">where conversations grow</p>
        </div>
        <div className="auth-tabs">
          <button className={`auth-tab ${isLogin ? "active" : ""}`} onClick={() => setIsLogin(true)}>Sign In</button>
          <button className={`auth-tab ${!isLogin ? "active" : ""}`} onClick={() => setIsLogin(false)}>Sign Up</button>
        </div>
        <div className="auth-form">
          {!isLogin && (
            <div className="field-group">
              <label className="field-label">Username</label>
              <input className="glass-input" type="text" placeholder="your_username" value={username} onChange={(e) => setUsername(e.target.value)} />
            </div>
          )}
          <div className="field-group">
            <label className="field-label">Email</label>
            <input className="glass-input" type="email" placeholder="you@example.com" value={email} onChange={(e) => setEmail(e.target.value)} />
          </div>
          <div className="field-group">
            <label className="field-label">Password</label>
            <input className="glass-input" type="password" placeholder="••••••••" value={password} onChange={(e) => setPassword(e.target.value)} onKeyDown={(e) => e.key === "Enter" && handleSubmit()} />
          </div>
          {error && <p className="auth-error">{error}</p>}
          <button className="btn-wood" onClick={handleSubmit} disabled={loading}>
            {loading ? "..." : isLogin ? "Enter the Grove" : "Plant Your Roots"}
          </button>
        </div>
      </div>
    </div>
  );
}