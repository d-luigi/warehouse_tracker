import type { Metadata } from "next";
import { Fredoka, Nunito } from "next/font/google";
import Scene from "@/components/Scene";
import "./globals.css";

const display = Fredoka({ subsets: ["latin"], variable: "--font-display" });
const body = Nunito({ subsets: ["latin"], variable: "--font-body" });

export const metadata: Metadata = {
  title: "Warehouse Tracker",
  description: "Track what is on your shelves, what came in and what went out.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${display.variable} ${body.variable}`}>
      <body>
        <Scene />
        {children}
      </body>
    </html>
  );
}
