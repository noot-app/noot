import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock env so Supabase SSR path runs without warnings
vi.mock('$env/dynamic/public', () => ({
  env: {
    PUBLIC_SUPABASE_URL: 'http://localhost:54321',
    PUBLIC_SUPABASE_ANON_KEY: 'anon'
  }
}));

// Minimal mock for createServerClient; we mostly assert locals wiring
const { mockCreateServerClient } = vi.hoisted(() => ({
  mockCreateServerClient: vi.fn().mockReturnValue({ auth: {} })
}));
vi.mock('@supabase/ssr', () => ({
  createServerClient: mockCreateServerClient
}));

// Mock getValidatedSession to control session presence
const { mockGetValidatedSession } = vi.hoisted(() => ({
  mockGetValidatedSession: vi.fn()
}));
vi.mock('$lib/utils.js', () => ({
  getValidatedSession: mockGetValidatedSession
}));

// Import after mocks
import { handle } from './hooks.server';

function makeEvent(pathname: string) {
  const cookies: any = {
    getAll: vi.fn().mockReturnValue([]),
    set: vi.fn()
  };
  const event: any = {
    url: new URL(`http://example.com${pathname}`),
    cookies,
    locals: {}
  };
  return event;
}

describe('hooks.server handle', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('allows public routes without session', async () => {
    mockGetValidatedSession.mockResolvedValueOnce(null);
    const event = makeEvent('/');
    const resolve = vi.fn().mockResolvedValue({ ok: true });

    const res = await handle({ event, resolve } as any);

    expect(mockCreateServerClient).toHaveBeenCalled();
    expect(typeof event.locals.getSession).toBe('function');
    expect(resolve).toHaveBeenCalledWith(event, expect.objectContaining({
      filterSerializedResponseHeaders: expect.any(Function)
    }));
    expect(res).toEqual({ ok: true });
  });

  it('redirects unauthenticated users from protected paths', async () => {
    mockGetValidatedSession.mockResolvedValueOnce(null);
    const event = makeEvent('/profile');
    const resolve = vi.fn();

    try {
      await handle({ event, resolve } as any);
      throw new Error('Expected redirect to be thrown');
    } catch (e: any) {
      expect(e.status).toBe(303);
      expect(decodeURIComponent(e.location)).toBe('/login?returnUrl=/profile');
    }
  });

  it('allows authenticated users on protected paths', async () => {
    mockGetValidatedSession.mockResolvedValueOnce({ access_token: 't', refresh_token: 'r', user: { id: 'u' } } as any);
    const event = makeEvent('/record/anything');
    const resolve = vi.fn().mockResolvedValue({ ok: true });

    const res = await handle({ event, resolve } as any);

    expect(resolve).toHaveBeenCalled();
    expect(res).toEqual({ ok: true });
  });

  it('exposes locals.getSession that returns the validated session', async () => {
    const fakeSession = { access_token: 't', refresh_token: 'r', user: { id: 'u' } } as any;
    // handle() calls getSession once internally; mock for all calls
    mockGetValidatedSession.mockResolvedValue(fakeSession);
    const event = makeEvent('/');
    const resolve = vi.fn().mockResolvedValue({ ok: true });

    await handle({ event, resolve } as any);
    const session = await event.locals.getSession();
    expect(session).toEqual(fakeSession);
  });

  it('filterSerializedResponseHeaders only allows specific headers', async () => {
    mockGetValidatedSession.mockResolvedValueOnce(null);
    const event = makeEvent('/');
    let optionsArg: any;
    const resolve = vi.fn().mockImplementation((_event: any, options?: any) => {
      optionsArg = options;
      return { ok: true };
    });

    await handle({ event, resolve } as any);

    expect(typeof optionsArg.filterSerializedResponseHeaders).toBe('function');
    const filter = optionsArg.filterSerializedResponseHeaders as (name: string) => boolean;
    expect(filter('content-range')).toBe(true);
    expect(filter('x-supabase-api-version')).toBe(true);
    expect(filter('set-cookie')).toBe(false);
  });
});
