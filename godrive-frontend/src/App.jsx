import { useState, useEffect } from "react";
import { IoSunnyOutline, IoMoonOutline } from "react-icons/io5";

import UploadForm from "./components/UploadForm";
import DownloadForm from "./components/DownloadForm";
import UpdateForm from "./components/UpdateForm";
import DeleteForm from "./components/DeleteForm";

import backgroundImageB from "./assets/background.avif";
import backgroundImageW from "./assets/backgroundW.avif";

export default function App() {
  const [darkMode, setDarkMode] = useState(true);

  useEffect(() => {
    const savedMode = localStorage.getItem("darkMode");
    if (savedMode !== null) setDarkMode(savedMode === "true");
  }, []);

  useEffect(() => {
    localStorage.setItem("darkMode", darkMode);
  }, [darkMode]);

  return (
    <div
      className="min-h-screen flex flex-col items-center py-12 px-6 transition-colors duration-500"
      style={{
        backgroundImage: darkMode ? `url(${backgroundImageB})` : `url(${backgroundImageW})`,
        backgroundSize: "cover",
        backgroundPosition: "center",
        backgroundRepeat: "no-repeat",
        color: darkMode ? "#f3f4f6" : "#111827",
      }}
    >
      <header className="w-full max-w-3xl flex justify-between items-center mb-10">
        <h1 className="text-3xl md:text-4xl font-semibold tracking-wide drop-shadow-md">
          GoDrive File Manager
        </h1>
        <button
          aria-label="Toggle Dark/Light Mode"
          onClick={() => setDarkMode(!darkMode)}
          className={`p-2 rounded-full focus:outline-none focus:ring-2 focus:ring-offset-2 transition-colors duration-300 ${
            darkMode
              ? "bg-gray-700 hover:bg-gray-600 focus:ring-gray-500"
              : "bg-gray-300 hover:bg-gray-400 focus:ring-gray-700"
          } flex items-center gap-2`}
        >
          {darkMode ? (
            <>
              <IoMoonOutline size={24} className="text-yellow-400" />
              <span className="sr-only">Dark mode enabled</span>
            </>
          ) : (
            <>
              <IoSunnyOutline size={24} className="text-yellow-500" />
              <span className="sr-only">Light mode enabled</span>
            </>
          )}
        </button>
      </header>

      <main className="w-full max-w-3xl grid grid-cols-1 md:grid-cols-2 gap-8">
        {[UploadForm, UpdateForm, DeleteForm, DownloadForm].map((FormComponent, i) => (
          <div
            key={i}
            className=" p-6 transform transition-transform duration-200 hover:scale-105 flex flex-col"
            style={{ minHeight: "360px" }}
          >
            <FormComponent darkMode={darkMode} />
          </div>
        ))}
      </main>

      <footer
        className={`mt-14 text-center text-xs select-none tracking-wide ${
          darkMode ? "text-gray-400" : "text-gray-600"
        }`}
      >
        &copy; 2025 GoDrive. All rights reserved.
      </footer>
    </div>
  );
}
