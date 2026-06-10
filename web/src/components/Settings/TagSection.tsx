import React, { useState, useEffect, useMemo } from "react";
import { AlertCircleIcon, Loader2Icon, TagIcon, Trash2Icon } from "lucide-react";
import { Button } from "@/components/ui/button";
import SettingTable from "@/components/Settings/SettingTable"; // Ensure path matches your project structure

interface TagDefinition {
  name: string;
  description: string;
}

export default function TagSection() {
  const [tags, setTags] = useState<TagDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  // Form States
  const [tagName, setTagName] = useState("");
  const [tagDescription, setTagDescription] = useState("");
  const [error, setError] = useState<string | null>(null);

  // Fetch tag definitions from your Go backend
  const fetchTags = async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/v1/tags");
      if (res.ok) {
        const data = await res.json();
        setTags(Array.isArray(data) ? data : []);
      } else {
        setError("Failed to sync tag definitions from server.");
      }
    } catch (err) {
      console.error(err);
      setError("Network error: Could not reach backend.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTags();
  }, []);

  // Real-time frontend restriction: No duplicate names allowed in system memory
  const duplicateTagName = useMemo(() => {
    const cleanInput = tagName.trim().toUpperCase();
    if (!cleanInput) return false;
    return tags.some(t => t.name.toUpperCase().trim() == cleanInput);
  }, [tagName, tags]);

  const isSubmitDisabled = submitting || !tagName.trim() || duplicateTagName;

  const handleCreateTag = async (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (isSubmitDisabled) return;

    setSubmitting(true);
    setError(null);

    try {
      const res = await fetch("/api/v1/tags", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: tagName.trim(),
          description: tagDescription.trim()
        })
      });

      if (res.ok) {
        setTagName("");
        setTagDescription("");
        fetchTags(); // Refresh the list
      } else {
        const errData = await res.json();
        setError(errData.error || "Failed to create tag.");
      }
    } catch (err) {
      setError("Failed to create tag due to a network connection error.");
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteTag = async (name: string) => {
    if (!window.confirm(`Are you sure you want to permanently delete the tag "${name}"?`)) return;

    try {
      const res = await fetch(`/api/v1/tags/${name}`, {
        method: "DELETE"
      });

      if (res.ok) {
        fetchTags();
      } else {
        const errData = await res.json();
        setError(errData.error || "Failed to delete tag.");
      }
    } catch (err) {
      setError("Network error: Could not process tag deletion.");
    }
  };

  // Table Column Schema Layout Blueprint Definitions
  const columns = [
    {
      key: "name",
      header: "Tag Name",
      className: "w-[30%]",
      render: (val: string) => (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-bold bg-slate-100 text-slate-800 border border-slate-200 uppercase tracking-wider">
          <TagIcon size={12} className="text-slate-500" />
          {val}
        </span>
      )
    },
    {
      key: "description",
      header: "Description",
      className: "w-[55%]",
      render: (val: string) => (
        <span className="text-slate-600 text-sm font-medium">
          {val || <span className="text-slate-400 italic font-normal">No description provided</span>}
        </span>
      )
    },
    {
      key: "actions",
      header: "",
      className: "w-[15%] text-right",
      render: (_: any, row: TagDefinition) => (
        <button
          type="button"
          onClick={() => handleDeleteTag(row.name)}
          className="flex items-center justify-center h-8 w-8 text-slate-400 hover:text-red-600 hover:bg-slate-100 rounded-lg transition-colors cursor-pointer"
        >
          <Trash2Icon size={18} />
        </button>
      )

    }
  ];

  return (
    <div className="space-y-6 max-w-4xl">
      <div>
        <h3 className="text-lg font-bold text-slate-800 flex items-center gap-2">
          <TagIcon size={20} className="text-slate-500" /> Tag Management
        </h3>
        <p className="text-sm text-slate-500 mt-0.5">
          Define global tags for system administrative use to classify, sort, and organize active clinical referrals.
        </p>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-xl text-sm font-medium flex items-center gap-2 animate-in fade-in-50 duration-150">
          <AlertCircleIcon size={16} className="text-red-500 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* CREATE NEW TAG PANEL WORKSPACE */}
      <div className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm">
        <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-4">Create Tag Definition</h4>
        <form onSubmit={handleCreateTag} className="flex flex-col md:flex-row gap-4 items-end">
          <div className="flex-1 w-full space-y-1.5">
            <label className="text-xs font-semibold text-slate-600">Tag Name</label>
            <input
              type="text"
              required
              placeholder="Enter tag name..."
              value={tagName}
              onChange={e => {
                setError(null);
                setTagName(e.target.value.replace(/\s+/g, "_")); // Keep format snake_case easily
              }}
              className="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm bg-white text-slate-900 outline-none focus:ring-2 focus:ring-blue-500/20"
            />
          </div>

          <div className="flex-2 w-full space-y-1.5">
            <label className="text-xs font-semibold text-slate-600">Description (Optional)</label>
            <input
              type="text"
              placeholder="Enter description..."
              value={tagDescription}
              onChange={e => setTagDescription(e.target.value)}
              className="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm bg-white text-slate-900 outline-none focus:ring-2 focus:ring-blue-500/20"
            />
          </div>

          <div className="w-full md:w-auto flex flex-col gap-1">
            {duplicateTagName && (
              <span className="text-[10px] text-red-500 font-bold mb-1 animate-in fade-in-50">
                Name already exists
              </span>
            )}
            <Button
              type="submit"
              disabled={isSubmitDisabled}
              className={`w-full md:w-auto ${isSubmitDisabled ? "bg-slate-400/20 text-slate-400 cursor-not-allowed shadow-none border-transparent" : ""}`}
            >
              {submitting ? "Adding..." : "Add Tag"}
            </Button>
          </div>
        </form>
      </div>

      {/* CENTRAL TAG MANAGEMENT DATAGRID LISTING */}
      <div className="mt-4">
        {loading ? (
          <div className="py-20 text-center text-sm text-slate-400 flex flex-col items-center gap-2">
            <Loader2Icon className="h-6 w-6 animate-spin text-blue-500" />
            <span>Syncing taxonomy definitions...</span>
          </div>
        ) : (
          <SettingTable
            columns={columns}
            data={tags}
            emptyMessage="No global tag identifiers initialized yet."
            className="border-slate-200 bg-transparent shadow-none"
            getRowKey={(row) => row.name}
          />
        )}
      </div>
    </div>
  );
}
