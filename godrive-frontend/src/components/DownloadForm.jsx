import { useState, useEffect } from "react";

export default function DownloadForm({ darkMode }) {
  const [fileName, setFileName] = useState("");
  const [fileContent, setFileContent] = useState("");
  const [message, setMessage] = useState(null); // { type: "success" | "error", text: string }

  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => setMessage(null), 5000);
    return () => clearTimeout(timer);
  }, [message]);

  const handleDownload = async () => {
    setMessage(null);
    try {
      const res = await fetch(`http://localhost:4000/download?filename=${fileName}`);
      if (!res.ok) throw new Error(await res.text());

      const data = await res.json();
      setFileContent(data.content || "");
      setMessage({ type: "success", text: "File downloaded successfully." });
    } catch (err) {
      setFileContent("");
      setMessage({ type: "error", text: "Download failed: " + err.message });
    }
  };

  const inputClass = `border p-2 w-full mb-2 rounded placeholder-opacity-70 ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 placeholder-neutral-400 border-neutral-600"
      : "bg-gray-200 text-gray-900 placeholder-neutral-900 border-neutral-900"
  }`;

  const textareaClass = `border p-3 w-full mt-2 rounded resize-none font-mono text-sm whitespace-pre-wrap break-words ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 placeholder-neutral-400 border-neutral-600"
      : "bg-gray-200 text-gray-900 placeholder-neutral-900 border-neutral-900"
  }`;

  const messageClass = message
    ? message.type === "success"
      ? "bg-lime-100 border border-lime-900 text-lime-700 px-4 py-2 rounded mt-2"
      : "bg-red-100 border border-red-400 text-red-700 px-4 py-2 rounded mt-2"
    : "";

  return (
    <div
      className={`p-4 rounded-xl shadow mb-4 ${
        darkMode ? "border border-neutral-600" : "border border-black"
      }`}
    >
      <h2 className="text-xl font-bold mb-2">Download File</h2>
      <input
        className={inputClass}
        placeholder="Filename"
        value={fileName}
        onChange={(e) => setFileName(e.target.value)}
      />
      <button
        onClick={handleDownload}
        className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 transition-colors"
      >
        Download
      </button>

      {message && <div className={messageClass}>{message.text}</div>}

      {fileContent !== "" && (
        <>
          <label
            htmlFor="fileContent"
            className={`block mt-4 mb-1 font-semibold ${
              darkMode ? "text-gray-300" : "text-gray-700"
            }`}
          >
            File Content
          </label>
          <textarea
            id="fileContent"
            className={textareaClass}
            rows="8"
            readOnly
            value={fileContent}
            placeholder="No content to display"
          />
        </>
      )}
    </div>
  );
}
