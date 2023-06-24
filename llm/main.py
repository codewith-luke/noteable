import os
from dotenv import load_dotenv
from langchain import OpenAI
from langchain.chains.question_answering import load_qa_chain
from langchain.chat_models import ChatOpenAI
from langchain.embeddings import OpenAIEmbeddings
from langchain.document_loaders import DirectoryLoader, UnstructuredMarkdownLoader
from langchain.vectorstores import Chroma

import updater

load_dotenv()

api_key = os.getenv("OPENAI_API_KEY")
# llm = OpenAI(openai_api_key=api_key, temperature=0, max_tokens=100)
llm = ChatOpenAI(openai_api_key=api_key, temperature=0, max_tokens=100)
embeddings = OpenAIEmbeddings(model="text-embedding-ada-002")


def insert_documents():
    loader = DirectoryLoader('./documents', glob="**/*.md")
    loader_docs = loader.load()
    ids = [loader_docs[i].metadata.get('source') for i in range(len(loader_docs))]
    conn = Chroma.from_documents(
        loader_docs,
        embeddings,
        collection_name="documents",
        persist_directory="./chroma_db",
        ids=ids
    )
    conn.persist()


def query_documents(message):
    db = Chroma(collection_name="documents", persist_directory='chroma_db', embedding_function=embeddings)
    docs = db.similarity_search(message, k=1)
    chain = load_qa_chain(llm, chain_type="stuff")
    res = chain.run(input_documents=docs, question=message)
    print(res)


def update_document(document_id):
    db = Chroma(collection_name="documents", persist_directory='chroma_db', embedding_function=embeddings)
    updater.update_doc()
    loader = UnstructuredMarkdownLoader(document_id)
    data = loader.load()
    db.update_document(document_id, data[0])
    db.persist()


# insert_documents()
query_documents("When is my next dentist appointment?")
# update_document("documents/test_document_3.md")
