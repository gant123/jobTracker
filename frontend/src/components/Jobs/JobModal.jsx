
import React, { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { JOB_STATUSES } from '../../utils/constants';

const defaultValues = {
  title: '',
  company: '',
  status: 'applied',
  location: '',
  link: '',
  salary: '',
  notes: '',
  appliedDate: '',
};

const JobModal = ({ job, onSave, onClose }) => {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm({ defaultValues });

  useEffect(() => {
    reset(job ? { ...defaultValues, ...job } : defaultValues);
  }, [job, reset]);

  const submit = (data) => onSave(data);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/30">
      <div className="w-full max-w-2xl bg-white rounded-2xl shadow-xl animate-slide-in">
        <div className="px-6 py-4 border-b flex items-center justify-between">
          <h3 className="text-xl font-semibold">
            {job ? 'Edit Job' : 'Add New Job'}
          </h3>
          <button className="btn-secondary" onClick={onClose}>Close</button>
        </div>

        <form onSubmit={handleSubmit(submit)} className="p-6 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium">Job Title</label>
              <input
                className={`input-field mt-1 ${errors.title ? 'border-red-500' : ''}`}
                placeholder="Software Engineer"
                {...register('title', { required: 'Title is required' })}
              />
              {errors.title && <p className="text-sm text-red-600 mt-1">{errors.title.message}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium">Company</label>
              <input
                className={`input-field mt-1 ${errors.company ? 'border-red-500' : ''}`}
                placeholder="Acme Inc."
                {...register('company', { required: 'Company is required' })}
              />
              {errors.company && <p className="text-sm text-red-600 mt-1">{errors.company.message}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium">Status</label>
              <select className="input-field mt-1" {...register('status')}>
                {JOB_STATUSES.map((s) => (
                  <option key={s} value={s}>
                    {s.charAt(0).toUpperCase() + s.slice(1)}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium">Location</label>
              <input className="input-field mt-1" placeholder="Remote / City, ST" {...register('location')} />
            </div>

            <div>
              <label className="block text-sm font-medium">Job Link</label>
              <input className="input-field mt-1" placeholder="https://..." {...register('link')} />
            </div>

            <div>
              <label className="block text-sm font-medium">Salary</label>
              <input className="input-field mt-1" placeholder="$120k–$140k" {...register('salary')} />
            </div>

            <div>
              <label className="block text-sm font-medium">Applied Date</label>
              <input type="date" className="input-field mt-1" {...register('appliedDate')} />
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium">Notes</label>
              <textarea
                rows={4}
                className="input-field mt-1"
                placeholder="Anything you want to remember..."
                {...register('notes')}
              />
            </div>
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" disabled={isSubmitting} className="btn-primary">
              {isSubmitting ? 'Saving...' : 'Save Job'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default JobModal;
