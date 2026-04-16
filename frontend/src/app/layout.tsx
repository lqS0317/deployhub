import type { Metadata } from "next";
import "./globals.css";
import { Providers } from "./providers";

export const metadata: Metadata = {
  title: "DeployHub",
  description: "K8s 原生运维发布平台",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh">
      <body className="min-h-screen bg-gray-50 antialiased">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
