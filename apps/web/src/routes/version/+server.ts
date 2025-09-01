import { json } from '@sveltejs/kit'
import type { RequestHandler } from './$types'

// These will be injected at build time by Vite
const buildInfo = {
  commit: import.meta.env.VITE_COMMIT_SHA || 'unknown',
  buildTime: import.meta.env.VITE_BUILD_TIME || 'unknown',
  tag: import.meta.env.VITE_TAG || 'dev'
}

export const GET: RequestHandler = async () => {
  return json({
    commit: buildInfo.commit,
    commitShort: buildInfo.commit !== 'unknown' && buildInfo.commit.length >= 7 
      ? buildInfo.commit.slice(0, 7) 
      : buildInfo.commit,
    buildTime: buildInfo.buildTime,
    tag: buildInfo.tag
  })
}