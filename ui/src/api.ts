export interface Doc {
	id?: string;
	title: string;
	path: string;
	content: string;
  updatedAt?: string;
  createdAt?: string
}

export interface User {
	id: string;
	email: string;
	name: string;
}

const basePath: string = "http://localhost:9001"
export async function login(email: string, password: string): Promise<void> {
	try {
		await fetch(`${basePath}/login`, {
			method: "POST",
			credentials: "include",
			body: JSON.stringify({email, token: password}),
		});
		console.log("Logged in!")
	} catch (err) {
		console.log(err);
	}
}

export async function getCurrentUser(): Promise<User> {
		const res = await fetch(`${basePath}/api/user`, {
			credentials: "include",
		});

		const user = await res.json();
		return user
}

export async function createDoc(doc: Doc): Promise<string> {
	const res = await fetch(`${basePath}/api/doc`, {
			method: "POST",
			credentials: "include",
			body: JSON.stringify(doc),
		}
	);

	const payload = await res.json();
	return payload.id;
}

export async function listDocs(): Promise<Doc[]> {
	try {
		const res = await fetch(`${basePath}/api/doc/list`, {
			credentials: "include",
		});
		const docs = await res.json();
		return docs;
	} catch (err) {
		console.log(err);
		return err;
	}
}

export async function getDoc(id: string): Promise<Doc> {
	try {
		const res = await fetch(`${basePath}/api/doc/${id}`, {
			credentials: "include",
		});
		const doc = await res.json();

		return doc;
	} catch (err) {
		console.log(err);
		return err;
	}
}

