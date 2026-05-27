import { useState, useEffect } from "react";
import { supabase } from "./lib/supabase";
import Auth from "./components/Auth/Auth";
import ChatApp from "./components/Chat/ChatApp";
import "./index.css";

export default function App() {
  const [session, setSession] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    supabase.auth.getSession().then(({ data: { session } }) => {
      setSession(session);
      setLoading(false);
    });

    const { data: { subscription } } = supabase.auth.onAuthStateChange((_event, session) => {
      setSession(session);
    });

    return () => subscription.unsubscribe();
  }, []);

  if (loading) return (
    <div className="splash">
      <div className="splash-knot">
        <div className="knot-ring" />
        <div className="knot-ring knot-ring--2" />
      </div>
      <p className="splash-text">Timber</p>
    </div>
  );

  return session ? <ChatApp session={session} /> : <Auth />;
}