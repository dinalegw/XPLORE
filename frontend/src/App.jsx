import { useState, useEffect } from "react";
import { supabase } from "./lib/supabase";
import Auth from "./components/Auth/Auth";
import ChatApp from "./components/Chat/ChatApp";
import { useChatStore } from "./store/chatStore";
import "./index.css";

export default function App() {
   const [session, setSession] = useState(null);
   const [loading, setLoading] = useState(true);
   const { wsError, clearWsError } = useChatStore();

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

   // Map technical errors to user-friendly messages
   const getUserFriendlyError = (error) => {
     if (!error) return null;
     
     const errorLower = error.toLowerCase();
     
     if (errorLower.includes('failed to fetch') || 
         errorLower.includes('network') ||
         errorLower.includes('connection')) {
       return "Unable to connect to the server. Please check your internet connection and try again.";
     }
     
     if (errorLower.includes('timeout')) {
       return "Connection timed out. Please try again.";
     }
     
     if (errorLower.includes('invalid or missing token') ||
         errorLower.includes('unauthorized') ||
         errorLower.includes('authentication')) {
       return "Authentication failed. Please sign in again.";
     }
     
     if (errorLower.includes('websocket')) {
       return "Connection interrupted. Please refresh the page to reconnect.";
     }
     
     // Default to a generic message for unknown errors
     return "An unexpected error occurred. Please try again.";
   };

   if (wsError) {
     const userFriendlyError = getUserFriendlyError(wsError);
     return (
       <div className="error-container">
         <p className="error-message">{userFriendlyError}</p>
         <button onClick={clearWsError}>Dismiss</button>
       </div>
     );
   }

   if (loading) return (
     <div className="splash">
       <div className="splash-knot">
         <div className="knot-ring" />
         <div className="knot-ring knot-ring--2" />
       </div>
       <p className="splash-text">XPLORE</p>
     </div>
   );

   return session ? <ChatApp session={session} /> : <Auth />;
 }
