
export function getCookieValue(allCookies: string, target: string): string | null {
	let keyStart = allCookies.indexOf(target);
	if (keyStart < 0) {
		return null;
	}
	let cookieDelim = keyStart + target.length;
	let delimIndex = allCookies.indexOf(";", cookieDelim);
	if (delimIndex < 0) { delimIndex = allCookies.length; }
	return allCookies.substring(cookieDelim + 1, delimIndex);
}
