export type ToastType = "success" | "error" | "info" | "warning";
export type ToastPosition = "top-left" | "top-center" | "top-right" | "bottom-left" | "bottom-center" | "bottom-right";

export interface Toast {
  id: string;
  message: string;
  type: ToastType;
  timeout: number;
  position: ToastPosition;
}