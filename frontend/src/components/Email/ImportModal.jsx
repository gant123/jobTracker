
import React, { useMemo, useState, useCallback, useEffect } from 'react';
import { XMarkIcon, ChevronDownIcon, ChevronRightIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

const JOB_STATUSES = ['wishlist', 'applied', 'interviewing', 'offer', 'rejected', 'withdrawn'];

const normalize = (e) => ({
  messageId: e.messageId || e.MessageID || e.message_id || '',
  company: e.company || e.Company || '',
  title: e.title || e.Title || e.position || e.Position || '',
  status: e.status || e.Status || 'applied',
  appliedDate: (e.appliedDate || e.AppliedDate || e.applied_date)
    ? new Date(e.appliedDate || e.AppliedDate || e.applied_date).toISOString().slice(0, 10)
    : new Date().toISOString().slice(0, 10),
  subject: e.subject || e.Subject || '',
  snippet: e.snippet || e.Snippet || '',
  link: e.link || e.Link || '',
  selected: true,
  open: false,
});

const Row = ({ index, rows, filtered, setRow }) => {
  const r = filtered[index];
  if (!r) return null;
  const idx = rows.indexOf(r);

  const statusColors = {
    wishlist: 'bg-gray-100 text-gray-800',
    applied: 'bg-blue-100 text-blue-800',
    interviewing: 'bg-yellow-100 text-yellow-800',
    offer: 'bg-green-100 text-green-800',
    rejected: 'bg-red-100 text-red-800',
    withdrawn: 'bg-gray-100 text-gray-800',
  };

  return (
    <div className="border-b border-gray-100 hover:bg-gray-50">
      <div className="grid grid-cols-[40px_1fr_1fr_120px_100px_2fr] gap-2 px-4 py-2 items-center">
        <input
          type="checkbox"
          checked={r.selected}
          onChange={(e) => setRow(idx, { selected: e.target.checked })}
          className="w-4 h-4 text-primary-600 rounded focus:ring-primary-500"
        />

        <input
          className="px-2 py-1 text-sm border rounded focus:outline-none focus:ring-1 focus:ring-primary-500"
          value={r.company}
          onChange={(e) => setRow(idx, { company: e.target.value })}
          placeholder="Company"
        />

        <input
          className="px-2 py-1 text-sm border rounded focus:outline-none focus:ring-1 focus:ring-primary-500"
          value={r.title}
          onChange={(e) => setRow(idx, { title: e.target.value })}
          placeholder="Position"
        />

        <select
          className={`px-2 py-1 text-xs font-medium rounded ${statusColors[r.status]}`}
          value={r.status}
          onChange={(e) => setRow(idx, { status: e.target.value })}
        >
          {JOB_STATUSES.map((s) => (
            <option key={s} value={s}>
              {s.charAt(0).toUpperCase() + s.slice(1)}
            </option>
          ))}
        </select>

        <input
          type="date"
          className="px-2 py-1 text-sm border rounded focus:outline-none focus:ring-1 focus:ring-primary-500"
          value={r.appliedDate}
          onChange={(e) => setRow(idx, { appliedDate: e.target.value })}
        />

        <div className="flex items-center gap-2 overflow-hidden">
          <button
            className="text-primary-600 hover:text-primary-700 text-xs font-medium flex items-center gap-1"
            onClick={() => setRow(idx, { open: !r.open })}
          >
            {r.open ? <ChevronDownIcon className="w-3 h-3" /> : <ChevronRightIcon className="w-3 h-3" />}
            Preview
          </button>

          {r.link && (
            <a
              className="text-primary-600 hover:text-primary-700 text-xs underline"
              href={r.link}
              target="_blank"
              rel="noreferrer"
            >
              Gmail
            </a>
          )}

          <div className="truncate text-xs text-gray-600" title={r.subject}>
            {r.subject}
          </div>
        </div>
      </div>

      {r.open && (
        <div className="px-4 pb-2 ml-10">
          <div className="bg-gray-50 rounded p-2 text-xs text-gray-600">
            {r.snippet || 'No preview available.'}
          </div>
        </div>
      )}
    </div>
  );
};

const ImportModal = ({ events = [], onClose, onImport }) => {
  const initial = useMemo(() => events.map(normalize), [events]);
  const [rows, setRows] = useState(initial);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [isImporting, setIsImporting] = useState(false);

  // auto-detect rejections on mount
  useEffect(() => {
    const rejectionKeywords = [
      'not moving forward',
      'unfortunately',
      'decided to move forward with other',
      'no longer being considered',
      'not selected',
      'pursue other candidates',
      'we regret',
    ];
    setRows((curr) =>
      curr.map((r) => {
        const s = (r.snippet || '').toLowerCase();
        const subj = (r.subject || '').toLowerCase();
        return rejectionKeywords.some((k) => s.includes(k) || subj.includes(k))
          ? { ...r, status: 'rejected' }
          : r;
      }),
    );
  }, []);

  const setRow = useCallback((idx, patch) => {
    setRows((rs) => rs.map((r, i) => (i === idx ? { ...r, ...patch } : r)));
  }, []);

  const filtered = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    return rows.filter((r) => {
      if (filterStatus !== 'all' && r.status !== filterStatus) return false;
      if (!q) return true;
      return (
        r.company.toLowerCase().includes(q) ||
        r.title.toLowerCase().includes(q) ||
        r.subject.toLowerCase().includes(q) ||
        r.snippet.toLowerCase().includes(q)
      );
    });
  }, [rows, searchQuery, filterStatus]);

  const selectedCount = rows.filter((r) => r.selected).length;
  const allSelected = filtered.length > 0 && filtered.every((r) => r.selected);

  const toggleAll = () => {
    const next = !allSelected;
    setRows((rs) => rs.map((r) => (filtered.includes(r) ? { ...r, selected: next } : r)));
  };

  const bulkUpdateStatus = (status) => {
    setRows((rs) => rs.map((r) => (r.selected ? { ...r, status } : r)));
    toast.success(`Marked ${selectedCount} items as ${status}`);
  };

  const selectByStatus = (status) => {
    setRows((rs) => rs.map((r) => ({ ...r, selected: r.status === status })));
  };

  const handleImport = async () => {
    const jobsToImport = rows
      .filter((r) => r.selected && (r.company || r.title))
      .map((r) => ({
        company: r.company.trim() || 'Unknown Company',
        position: r.title.trim() || 'Unknown Position',
        status: r.status,
        applied_date: r.appliedDate ? `${r.appliedDate}T00:00:00Z` : undefined,
        url: r.link || undefined,
        notes: `[Imported from Gmail] ${r.subject}`,
        location: '',
        gmail_message_id: r.messageId
      }));
    console.log('Jobs to import:', JSON.stringify(jobsToImport, null, 2));
    if (jobsToImport.length === 0) {
      toast.error('Please select at least one job to import');
      return;
    }

    setIsImporting(true);
    try {
      await onImport(jobsToImport);
      onClose();
    } catch (e) {
      console.error(e);
      toast.error('Failed to import some jobs');
    } finally {
      setIsImporting(false);
    }
  };

  if (!events || events.length === 0) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      {/* min-h-0 so inner section can scroll */}
      <div className="w-full max-w-7xl max-h-[90vh] bg-white rounded-xl shadow-2xl flex flex-col min-h-0">
        {/* Header */}
        <div className="px-6 py-4 border-b flex items-center justify-between bg-gradient-to-r from-primary-50 to-white">
          <div>
            <h3 className="text-xl font-bold text-gray-900">Import from Gmail</h3>
            <p className="text-sm text-gray-600 mt-1">
              Found {events.length} job-related emails. Review and import selected items.
            </p>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        {/* Toolbar */}
        <div className="px-6 py-3 border-b bg-gray-50 flex items-center gap-3 flex-wrap">
          <input
            className="flex-1 min-w-[200px] px-3 py-1.5 text-sm border rounded focus:outline-none focus:ring-1 focus:ring-primary-500"
            placeholder="Search emails..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <select
            className="px-3 py-1.5 text-sm border rounded focus:outline-none focus:ring-1 focus:ring-primary-500"
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
          >
            <option value="all">All Status</option>
            <option value="applied">Applied</option>
            <option value="rejected">Rejected</option>
            <option value="interviewing">Interviewing</option>
          </select>
          <div className="flex gap-2 ml-auto">
            <button
              className="px-3 py-1.5 text-sm bg-red-50 text-red-600 rounded hover:bg-red-100"
              onClick={() => selectByStatus('rejected')}
            >
              Select Rejected
            </button>
            <button
              className="px-3 py-1.5 text-sm bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
              onClick={() => bulkUpdateStatus('rejected')}
              disabled={selectedCount === 0}
            >
              Mark as Rejected
            </button>
          </div>
        </div>

        {/* Main area (header row + scroll area) */}
        <div className="flex-1 min-h-0 flex flex-col">
          {/* Column headers */}
          <div className="px-4 py-2 border-b bg-gray-50">
            <div className="grid grid-cols-[40px_1fr_1fr_120px_100px_2fr] gap-2 text-xs font-medium text-gray-500 uppercase tracking-wider">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={allSelected}
                  onChange={toggleAll}
                  className="w-4 h-4 text-primary-600 rounded"
                />
              </label>
              <div>Company</div>
              <div>Position</div>
              <div>Status</div>
              <div>Date</div>
              <div>Email Subject</div>
            </div>
          </div>

          {/* Scrollable list (native scroll, no virtualization) */}
          <div className="flex-1 min-h-0 overflow-y-auto">
            {filtered.map((_, i) => (
              <Row
                key={filtered[i].messageId || i}
                index={i}
                rows={rows}
                filtered={filtered}
                setRow={setRow}
              />
            ))}
          </div>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t bg-gray-50 flex items-center justify-between">
          <div className="text-sm text-gray-600">
            <span className="font-medium">{selectedCount}</span> selected ·
            <span className="font-medium ml-2">{filtered.length}</span> shown ·
            <span className="font-medium ml-2">{rows.length}</span> total
          </div>
          <div className="flex gap-3">
            <button
              className="px-4 py-2 text-sm bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
              onClick={onClose}
              disabled={isImporting}
            >
              Cancel
            </button>
            <button
              className="px-4 py-2 text-sm bg-primary-600 text-white rounded hover:bg-primary-700 disabled:opacity-50 min-w-[120px]"
              onClick={handleImport}
              disabled={selectedCount === 0 || isImporting}
            >
              {isImporting ? 'Importing...' : `Import ${selectedCount}`}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ImportModal;

