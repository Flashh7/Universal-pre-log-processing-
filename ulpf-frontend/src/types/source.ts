export type SourceStatus =
  | "Active"
  | "Inactive"
  | "Error";

export interface Source {
  source_uid: string;
  name: string;
  type: string;
  vendor_name: string;
  product_name: string;
  original_format: string;
  status: SourceStatus;
}