import { useState } from "react";
import { createSource } from "../api/sources";
import type { SourceStatus } from "../types/source";
import type { Page } from "../components/layout";

interface AddSourceProps {
  setActivePage: (page: Page) => void;
}

function AddSource({ setActivePage }: AddSourceProps) {
  const [formData, setFormData] = useState({
    name: "",
    type: "",
    vendor_name: "",
    product_name: "",
    original_format: "",
    status: "Active" as SourceStatus,
  });

  const [customSourceType, setCustomSourceType] = useState(false);
  const [customInputFormat, setCustomInputFormat] = useState(false);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;

    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSourceTypeChange = (
    e: React.ChangeEvent<HTMLSelectElement>
  ) => {
    const value = e.target.value;

    if (value === "__custom__") {
      setCustomSourceType(true);

      setFormData((prev) => ({
        ...prev,
        type: "",
      }));

      return;
    }

    setCustomSourceType(false);

    setFormData((prev) => ({
      ...prev,
      type: value,
    }));
  };

  const handleInputFormatChange = (
    e: React.ChangeEvent<HTMLSelectElement>
  ) => {
    const value = e.target.value;

    if (value === "__custom__") {
      setCustomInputFormat(true);

      setFormData((prev) => ({
        ...prev,
        original_format: "",
      }));

      return;
    }

    setCustomInputFormat(false);

    setFormData((prev) => ({
      ...prev,
      original_format: value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    setError("");
    setSubmitting(true);

    try {
      await createSource(formData);

      // Return to Sources after successful creation
      setActivePage("sources");
    } catch (error) {
      console.error("Failed to create source:", error);

      setError(
        error instanceof Error
          ? error.message
          : "Failed to create source"
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="add-source-page">

      <div className="add-source-header">
        <h1>Add Source</h1>
        <p>Connect a new log source to ULPF.</p>
      </div>

      <form onSubmit={handleSubmit}>

        <div className="form-section">
          <h2>Source Information</h2>

          <div className="form-grid">

            <div className="form-field">
              <label htmlFor="name">
                Source Name
              </label>

              <input
                id="name"
                type="text"
                name="name"
                value={formData.name}
                onChange={handleChange}
                placeholder="e.g. Firewall-01"
                required
              />
            </div>

            <div className="form-field">
              <label htmlFor="vendor_name">
                Vendor
              </label>

              <input
                id="vendor_name"
                type="text"
                name="vendor_name"
                value={formData.vendor_name}
                onChange={handleChange}
                placeholder="e.g. Cisco"
                required
              />
            </div>

            <div className="form-field">
              <label htmlFor="product_name">
                Product
              </label>

              <input
                id="product_name"
                type="text"
                name="product_name"
                value={formData.product_name}
                onChange={handleChange}
                placeholder="e.g. ASA Firewall"
                required
              />
            </div>

            <div className="form-field">
              <label htmlFor="type">
                Source Type
              </label>

              {!customSourceType ? (
                <select
                  id="type"
                  name="type"
                  value={formData.type}
                  onChange={handleSourceTypeChange}
                  required
                >
                  <option value="">
                    Select source type
                  </option>

                  <option value="Firewall">
                    Firewall
                  </option>

                  <option value="Server">
                    Server
                  </option>

                  <option value="Endpoint">
                    Endpoint
                  </option>

                  <option value="Network">
                    Network
                  </option>

                  <option value="Application">
                    Application
                  </option>

                  <option value="Database">
                    Database
                  </option>

                  <option value="__custom__">
                    + Add new type
                  </option>
                </select>
              ) : (
                <input
                  id="type"
                  type="text"
                  name="type"
                  value={formData.type}
                  onChange={handleChange}
                  placeholder="Enter new source type"
                  required
                />
              )}
            </div>

            <div className="form-field">
              <label htmlFor="original_format">
                Input Format
              </label>

              {!customInputFormat ? (
                <select
                  id="original_format"
                  name="original_format"
                  value={formData.original_format}
                  onChange={handleInputFormatChange}
                  required
                >
                  <option value="">
                    Select format
                  </option>

                  <option value="Syslog">
                    Syslog
                  </option>

                  <option value="JSON">
                    JSON
                  </option>

                  <option value="XML">
                    XML
                  </option>

                  <option value="CSV">
                    CSV
                  </option>

                  <option value="CEF">
                    CEF
                  </option>

                  <option value="LEEF">
                    LEEF
                  </option>

                  <option value="__custom__">
                    + Add new format
                  </option>
                </select>
              ) : (
                <input
                  id="original_format"
                  type="text"
                  name="original_format"
                  value={formData.original_format}
                  onChange={handleChange}
                  placeholder="Enter new log format"
                  required
                />
              )}
            </div>

            <div className="form-field">
              <label htmlFor="status">
                Status
              </label>

              <select
                id="status"
                name="status"
                value={formData.status}
                onChange={handleChange}
                required
              >
                <option value="Active">
                  Active
                </option>

                <option value="Inactive">
                  Inactive
                </option>

                <option value="Error">
                  Error
                </option>
              </select>
            </div>

          </div>
        </div>

        {error && (
          <div className="form-error">
            {error}
          </div>
        )}

        <div className="form-actions">

          <button
            type="button"
            onClick={() => setActivePage("sources")}
            disabled={submitting}
          >
            Cancel
          </button>

          <button
            type="submit"
            disabled={submitting}
          >
            {submitting ? "Adding..." : "Add Source"}
          </button>

        </div>

      </form>

    </div>
  );
}

export default AddSource;

