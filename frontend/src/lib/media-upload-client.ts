import { getToken } from '$lib/api/client';
import type { components } from '$lib/api/types';
import { getApiBase } from '$lib/stores/instance.svelte';

export type MediaUploadResult = components['schemas']['MediaUploadResult'];

interface UploadMediaFileOptions {
	workspaceId: string;
	file: File;
	altText?: string;
}

interface UploadProblem {
	detail?: string;
	error?: string;
	title?: string;
}

export class UploadRequestError extends Error {
	status: number;

	constructor(message: string, status: number) {
		super(message);
		this.name = 'UploadRequestError';
		this.status = status;
	}
}

export async function uploadMediaFile({
	workspaceId,
	file,
	altText = ''
}: UploadMediaFileOptions): Promise<MediaUploadResult> {
	try {
		return await uploadViaDirectSession(workspaceId, file, altText);
	} catch (error) {
		if (!shouldUseMultipartFallback(error)) {
			throw error;
		}
		return uploadViaMultipart(workspaceId, file, altText);
	}
}

export async function uploadMediaFiles(
	workspaceId: string,
	files: FileList | File[],
	onProgress?: (uploaded: number, total: number) => void
): Promise<MediaUploadResult[]> {
	const selectedFiles = Array.from(files).filter(isSupportedMediaFile);
	const results: MediaUploadResult[] = [];
	for (const file of selectedFiles) {
		results.push(await uploadMediaFile({ workspaceId, file }));
		onProgress?.(results.length, selectedFiles.length);
	}
	return results;
}

export function isSupportedMediaFile(file: File): boolean {
	return file.type.startsWith('image/') || file.type.startsWith('video/');
}

export function shouldUseMultipartFallback(error: unknown): boolean {
	if (!(error instanceof UploadRequestError)) {
		return false;
	}
	if (error.status === 404 || error.status === 405) {
		return true;
	}
	return error.message.toLowerCase().includes('direct media upload sessions require s3 storage');
}

export function directUploadHeadersForBrowser(headers: Record<string, string>): Headers {
	const directHeaders = new Headers();
	for (const [key, value] of Object.entries(headers)) {
		if (!value || isForbiddenBrowserUploadHeader(key)) {
			continue;
		}
		directHeaders.set(key, value);
	}
	return directHeaders;
}

async function uploadViaDirectSession(
	workspaceId: string,
	file: File,
	altText: string
): Promise<MediaUploadResult> {
	const sessionResp = await fetch(apiURL('/media/upload-session'), {
		method: 'POST',
		headers: apiHeaders(true),
		body: JSON.stringify({
			workspace_id: workspaceId,
			filename: file.name,
			mime_type: file.type || 'application/octet-stream',
			size: file.size,
			...(altText ? { alt_text: altText } : {})
		})
	});
	if (!sessionResp.ok) {
		throw await uploadErrorFromResponse(sessionResp, 'Failed to create upload session');
	}

	const session =
		(await sessionResp.json()) as components['schemas']['CreateMediaUploadSessionOutputBody'];
	const uploadHeaders = directUploadHeadersForBrowser(session.upload.headers ?? {});
	if (!uploadHeaders.has('Content-Type') && file.type) {
		uploadHeaders.set('Content-Type', file.type);
	}
	const uploadResp = await fetch(session.upload.url, {
		method: session.upload.method || 'PUT',
		headers: uploadHeaders,
		body: file
	});
	if (!uploadResp.ok) {
		throw await uploadErrorFromResponse(uploadResp, 'Direct media upload failed');
	}

	const completeResp = await fetch(apiURL(session.complete_url), {
		method: 'POST',
		headers: apiHeaders(true),
		body: JSON.stringify({ workspace_id: workspaceId })
	});
	if (!completeResp.ok) {
		throw await uploadErrorFromResponse(completeResp, 'Failed to finalize media upload');
	}
	return (await completeResp.json()) as MediaUploadResult;
}

async function uploadViaMultipart(
	workspaceId: string,
	file: File,
	altText: string
): Promise<MediaUploadResult> {
	const formData = new FormData();
	formData.append('file', file);
	formData.append('workspace_id', workspaceId);
	if (altText) {
		formData.append('alt_text', altText);
	}

	const response = await fetch(apiURL('/media/upload'), {
		method: 'POST',
		headers: apiHeaders(false),
		body: formData
	});
	if (!response.ok) {
		throw await uploadErrorFromResponse(response, 'Upload failed');
	}
	return (await response.json()) as MediaUploadResult;
}

function apiURL(path: string): string {
	if (path.startsWith('http://') || path.startsWith('https://')) {
		return path;
	}
	const apiPath = path.startsWith('/api/v1/') ? path.slice('/api/v1'.length) : path;
	return `${getApiBase()}${apiPath.startsWith('/') ? apiPath : `/${apiPath}`}`;
}

function apiHeaders(json: boolean): Headers {
	const headers = new Headers();
	if (json) {
		headers.set('Content-Type', 'application/json');
	}
	const token = getToken();
	if (token) {
		headers.set('Authorization', `Bearer ${token}`);
	}
	return headers;
}

async function uploadErrorFromResponse(
	response: Response,
	fallback: string
): Promise<UploadRequestError> {
	const problem = await parseUploadProblem(response);
	const message =
		problem.detail || problem.error || problem.title || `${fallback} (${response.status})`;
	return new UploadRequestError(message, response.status);
}

async function parseUploadProblem(response: Response): Promise<UploadProblem> {
	const contentType = response.headers.get('Content-Type') ?? '';
	if (!contentType.includes('json')) {
		return { detail: await response.text() };
	}
	try {
		return (await response.json()) as UploadProblem;
	} catch {
		return {};
	}
}

function isForbiddenBrowserUploadHeader(header: string): boolean {
	const normalized = header.toLowerCase();
	return normalized === 'host' || normalized === 'content-length';
}
