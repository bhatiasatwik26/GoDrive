import { useState, useEffect } from "react";

export default function UploadForm({ darkMode }) {
  const [fileName, setFileName] = useState("");
  const [content, setContent] = useState("");
  const [message, setMessage] = useState(null); // { type: "success" | "error", text: string }

  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => setMessage(null), 5000);
    return () => clearTimeout(timer);
  }, [message]);

  const handleUpload = async () => {
    setMessage(null);
    try {
      const res = await fetch("http://localhost:4000/upload", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ fileName, content }),
      });
      const text = await res.text();
      if (!res.ok) {
        setMessage({ type: "error", text: "Upload failed: " + text });
      } else {
        setMessage({ type: "success", text });
      }
    } catch (err) {
      setMessage({ type: "error", text: "Upload failed: " + err.message });
    }
  };

  const inputClass = `border p-2 w-full mb-2 rounded placeholder-opacity-70 ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 placeholder-neutral-400 border-neutral-600"
      : "bg-gray-200 text-gray-900 placeholder-neutral-900 border-neutral-900"
  }`;

  const textareaClass = `border p-2 w-full mb-2 rounded placeholder-opacity-70 resize-none ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 placeholder-neutral-400 border-neutral-600"
      : "bg-gray-200 text-gray-900 placeholder-neutral-900 border-neutral-900"
  }`;

  const messageClass =
    message?.type === "success"
      ? "bg-lime-100 border border-lime-900 text-lime-700 px-4 py-2 rounded mt-2"
      : "bg-red-100 border border-red-400 text-red-700 px-4 py-2 rounded mt-2";

  return (
    <div
      className={`p-4 rounded-xl shadow mb-4 ${
        darkMode ? "border border-neutral-600" : "border border-black"
      }`}
    >
      <h2 className="text-xl font-bold mb-2">Upload File</h2>
      <input
        className={inputClass}
        placeholder="Filename"
        value={fileName}
        onChange={(e) => setFileName(e.target.value)}
      />
      <textarea
        className={textareaClass}
        rows="4"
        placeholder="File content"
        value={content}
        onChange={(e) => setContent(e.target.value)}
      />
      <button
        onClick={handleUpload}
        className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 transition-colors"
      >
        Upload
      </button>
      {message && <div className={messageClass}>{message.text}</div>}
    </div>
  );
}
