import { useState, useEffect } from "react";

export default function UpdateForm({ darkMode }) {
  const [fileName, setFileName] = useState("");
  const [content, setContent] = useState("");
  const [message, setMessage] = useState(null); // { type: "success" | "error", text: string }

  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => setMessage(null), 5000);
    return () => clearTimeout(timer);
  }, [message]);

  const handleUpdate = async () => {
    setMessage(null);
    try {
      const res = await fetch("http://localhost:4000/update", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ fileName, content }),
      });
      if (!res.ok) throw new Error(await res.text());
      const msg = await res.text();
      setMessage({ type: "success", text: msg });
    } catch (err) {
      setMessage({ type: "error", text: "Update failed: " + err.message });
    }
  };

  const inputClass = `border p-2 w-full mb-2 rounded placeholder-opacity-70 ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 placeholder-neutral-400 border-neutral-600"
      : "bg-gray-200 text-gray-900 placeholder-neutral-900 border-neutral-900"
  }`;

  const textareaClass = `border p-2 w-full mb-2 rounded resize-none placeholder-opacity-70 ${
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
      <h2 className="text-xl font-bold mb-2">Update File</h2>
      <input
        className={inputClass}
        placeholder="Filename"
        value={fileName}
        onChange={(e) => setFileName(e.target.value)}
      />
      <textarea
        className={textareaClass}
        rows="4"
        placeholder="New content"
        value={content}
        onChange={(e) => setContent(e.target.value)}
      />
      <button
        onClick={handleUpdate}
        className="bg-yellow-500 text-white px-4 py-2 rounded hover:bg-yellow-600 transition-colors"
      >
        Update
      </button>

      {message && <div className={messageClass}>{message.text}</div>}
    </div>
  );
}
