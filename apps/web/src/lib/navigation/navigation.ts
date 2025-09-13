import MicrophoneIcon from "$lib/components/icons/Microphone.svelte"
import ChartBarIcon from "$lib/components/icons/ChartBar.svelte"
import SquaresTwoByTwo from "$lib/components/icons/SquaresTwoByTwo.svelte"
import TimelineIcon from "$lib/components/icons/Timeline.svelte"
import CalendarIcon from "$lib/components/icons/calendar-days.svelte"
import TagIcon from "$lib/components/icons/Tag.svelte"
import UserIcon from "$lib/components/icons/User.svelte"
import CursorArrowRaysIcon from "$lib/components/icons/CursorArrowRays.svelte"
import MenuIcon from "$lib/components/icons/Menu.svelte"

export const iconComponents = {
  microphone: MicrophoneIcon,
  "chart-bar": ChartBarIcon,
  dashboard: SquaresTwoByTwo,
  timeline: TimelineIcon,
  calendar: CalendarIcon,
  tag: TagIcon,
  user: UserIcon,
  "cursor-arrow-rays": CursorArrowRaysIcon,
  menu: MenuIcon,
}

export type NavIcon = keyof typeof iconComponents

export interface NavItemBase {
  id: string
  label: string
  icon: NavIcon
  href?: string // optional for trigger items
  type: "route" | "trigger"
}

// Core route items (canonical definitions)
export const routeItems: NavItemBase[] = [
  { id: "record", label: "Record", icon: "microphone", href: "/record", type: "route" },
  { id: "quick", label: "Quick", icon: "cursor-arrow-rays", href: "/quick", type: "route" },
  { id: "summary", label: "Summary", icon: "chart-bar", href: "/summary", type: "route" },
  { id: "dashboard", label: "Dashboard", icon: "dashboard", href: "/dashboard", type: "route" },
  { id: "log", label: "Log", icon: "timeline", href: "/log", type: "route" },
  { id: "events", label: "Events", icon: "calendar", href: "/events", type: "route" },
  { id: "labels", label: "Labels", icon: "tag", href: "/labels", type: "route" },
  { id: "profile", label: "Profile", icon: "user", href: "/profile", type: "route" },
]

// Bottom tab bar (direct visible tabs) — intentionally minimal (4 primary routes)
// Fifth item on the bar is the menu trigger (defined separately below)
export const bottomTabItems: NavItemBase[] = [
  routeItems.find(r => r.id === "record")!,
  routeItems.find(r => r.id === "summary")!,
  routeItems.find(r => r.id === "log")!,
  routeItems.find(r => r.id === "profile")!,
]

// Trigger item for bottom drawer
export const bottomMenuTrigger: NavItemBase = {
  id: "menu-trigger",
  label: "Menu",
  icon: "menu",
  type: "trigger",
}

// Items inside the bottom drawer (comprehensive: all routes)
export const drawerNavItems: NavItemBase[] = [
  routeItems.find(r => r.id === "record")!,
  routeItems.find(r => r.id === "quick")!,
  routeItems.find(r => r.id === "summary")!,
  routeItems.find(r => r.id === "dashboard")!,
  routeItems.find(r => r.id === "log")!,
  routeItems.find(r => r.id === "events")!,
  routeItems.find(r => r.id === "labels")!,
  routeItems.find(r => r.id === "profile")!,
]

// Top (desktop) navbar full set — retains all primary routes
export const topNavItems: NavItemBase[] = [
  routeItems.find(r => r.id === "record")!,
  routeItems.find(r => r.id === "quick")!,
  routeItems.find(r => r.id === "summary")!,
  routeItems.find(r => r.id === "dashboard")!,
  routeItems.find(r => r.id === "log")!,
  routeItems.find(r => r.id === "events")!,
  routeItems.find(r => r.id === "labels")!,
  routeItems.find(r => r.id === "profile")!,
]

export function getIconComponent(icon: NavIcon) {
  return iconComponents[icon]
}
