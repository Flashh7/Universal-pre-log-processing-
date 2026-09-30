export interface Source {
  source_uid: string;
  source_name: string;
  vendor_name: string;
  source_type: string;
  input_format: string;
  status: "Active" | "Inactive" | "Error";
  event_count: number;
}